package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/payment-processor/internal/debit/application/ports"
	"github.com/payment-processor/internal/debit/domain"
	"github.com/payment-processor/internal/debit/infra/repository"
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
	slog.InfoContext(ctx, "Handling request for user %s", req.UserID)

	var err error
	var wallet domain.Wallet

	for i := 0; i < maxRetries; i++ {
		wallet, err = h.walletRepo.Get(ctx, req.UserID)
		if err != nil {
			slog.ErrorContext(ctx, "error getting funds for user %s", req.UserID)
			return domain.NewGetFundsError(string(req.UserID), err) // Error no recuperable, salimos.
		}

		if err = wallet.Debit(req.Amount); err != nil {
			slog.ErrorContext(ctx, "error debiting amount for user", "amount", req.Amount, "userId", req.UserID)
			return err
		}

		err = h.walletRepo.Update(ctx, wallet)
		if err == nil {
			slog.InfoContext(ctx, "Debited amount for user %s", req.UserID)
			break
		}

		// Optimistic blocking
		if errors.Is(err, repository.ErrVersionMismatch) {
			slog.WarnContext(ctx, "version mismatch detected, retrying transaction", "attempt", i+1, "userId", req.UserID)
			continue
		}

		slog.ErrorContext(ctx, "unrecoverable repository error on update", "error", err, "userId", req.UserID)
		return domain.NewDebitFundsError(string(req.UserID), err)
	}

	if err != nil {
		slog.ErrorContext(ctx, "transaction failed after max retries", "error", err, "userId", req.UserID)
		return domain.NewMaxRetriesError(string(req.UserID), err)
	}

	if err = h.eventProcessor.Publish(ctx, toDebitEventRequest(wallet, req.Amount)); err != nil {
		slog.ErrorContext(ctx, "error publishing event after successful debit", "error", err)
		return domain.NewPublishMessageError(string(req.UserID), err)
	}

	slog.InfoContext(ctx, "Finished request for user %s", req.UserID)
	return nil
}

func (h *UseCaseHandler) userCanWithdraw(wantToWithdraw, canWithdraw domain.Amount) bool {
	return canWithdraw >= wantToWithdraw
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
