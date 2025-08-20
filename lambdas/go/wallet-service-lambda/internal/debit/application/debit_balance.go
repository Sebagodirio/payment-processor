package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/payment-processor/internal/debit/application/ports"
	"github.com/payment-processor/internal/debit/domain"
	"github.com/payment-processor/internal/debit/infra/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const maxRetries = 3

type (
	Request struct {
		UserID        domain.UserID
		Amount        domain.Amount
		CorrelationID string
	}

	UseCaseHandler struct {
		walletRepo     ports.WalletRepository
		eventProcessor ports.EventBusProcessor
	}
)

func (h *UseCaseHandler) Handle(ctx context.Context, req Request) error {
	tracer := otel.Tracer("wallet-service.application")
	ctx, span := tracer.Start(ctx, "UseCase.HandleDebit")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", string(req.UserID)),
		attribute.Float64("debit.amount", float64(req.Amount)),
	)

	slog.InfoContext(ctx, "Handling debit request", "userID", req.UserID)

	var err error
	var wallet domain.Wallet

	for i := 0; i < maxRetries; i++ {
		readCtx, readSpan := tracer.Start(ctx, "Repository.Get")
		wallet, err = h.walletRepo.Get(readCtx, req.UserID)
		readSpan.End()

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to get wallet")
			slog.ErrorContext(ctx, "Error getting funds for user", "userID", req.UserID, "error", err)
			return domain.NewGetFundsError(string(req.UserID), err) // Error no recuperable, salimos.
		}

		if err = wallet.Debit(req.Amount); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Insufficient funds")
			slog.ErrorContext(ctx, "Error debiting amount from wallet", "amount", req.Amount, "userID", req.UserID, "error", err)
			return err
		}

		updateCtx, updateSpan := tracer.Start(ctx, "Repository.UpdateWithOutbox")
		err = h.walletRepo.Update(updateCtx, wallet)
		updateSpan.End()

		if err == nil {
			slog.InfoContext(ctx, "Debited amount for user %s", "userID", req.UserID)
			break
		}

		// Optimistic blocking
		if errors.Is(err, repository.ErrVersionMismatch) {
			slog.WarnContext(ctx, "version mismatch detected, retrying transaction", "attempt", i+1, "userId", req.UserID)
			continue
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "Unrecoverable repository error")
		slog.ErrorContext(ctx, "unrecoverable repository error on update", "error", err, "userId", req.UserID)
		return domain.NewDebitFundsError(string(req.UserID), err)
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Transaction failed after max retries")
		slog.ErrorContext(ctx, "transaction failed after max retries", "error", err, "userId", req.UserID)
		return domain.NewMaxRetriesError(string(req.UserID), err)
	}

	if err = h.eventProcessor.Publish(ctx, toDebitEventRequest(wallet, req.Amount)); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Publish event failed")
		slog.ErrorContext(ctx, "error publishing event after successful debit", "error", err)
		return domain.NewPublishMessageError(string(req.UserID), err)
	}

	slog.InfoContext(ctx, "Finished request for user %s", "userID", req.UserID)
	return nil
}

func toDebitEventRequest(wallet domain.Wallet, amountToDebit domain.Amount) ports.BalanceDebitedRequest {
	return ports.BalanceDebitedRequest{
		UserID:        wallet.UserID,
		AmountDebited: amountToDebit,
		AmountLeft:    wallet.Amount,
		EventName:     domain.BalanceDebitedEventName,
	}
}

func NewDebitBalanceUseCaseHandler(repo ports.WalletRepository, bus ports.EventBusProcessor) *UseCaseHandler {
	return &UseCaseHandler{
		walletRepo:     repo,
		eventProcessor: bus,
	}
}
