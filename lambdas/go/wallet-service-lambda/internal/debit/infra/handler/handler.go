package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/payment-processor/internal/debit/application"
	events2 "github.com/payment-processor/internal/debit/domain/events"
)

var ErrValidation = errors.New("event validation failed")

type UseCase interface {
	Handle(ctx context.Context, req application.Request) error
}

type SQSHandler struct {
	useCase UseCase
}

func (h *SQSHandler) Handle(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		if err := h.processMessage(ctx, message); err != nil {
			slog.ErrorContext(
				ctx,
				"error processing message, batch will be retried",
				"messageId", message.MessageId,
				"error", err,
			)
			return err
		}
	}

	return nil
}

func (h *SQSHandler) processMessage(ctx context.Context, message events.SQSMessage) error {
	slog.InfoContext(ctx, "Processing SQS message", "messageId", message.MessageId)

	var event events2.PaymentInitEvent
	if err := json.Unmarshal([]byte(message.Body), &event); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal message body", "error", err, "body", message.Body)
		return err
	}

	logger := slog.With("correlationId", event.Header.CorrelationID)

	if err := h.validate(event); err != nil {
		logger.ErrorContext(ctx, "event validation failed", "error", err)
		return nil
	}

	if err := h.useCase.Handle(ctx, toUseCaseRequest(event.Payload, event.Header.CorrelationID)); err != nil {
		logger.ErrorContext(ctx, "use case failed to handle request", "error", err)
		return err
	}

	logger.InfoContext(ctx, "Successfully processed message", "messageId", message.MessageId)
	return nil
}

func (h *SQSHandler) validate(event events2.PaymentInitEvent) error {
	if event.Header.CorrelationID == "" {
		return errors.Join(ErrValidation, errors.New("correlation_id is missing"))
	}
	if event.Payload.UserID == "" {
		return errors.Join(ErrValidation, errors.New("user_id is missing"))
	}
	if event.Payload.Amount <= 0 {
		return errors.Join(ErrValidation, errors.New("amount must be positive"))
	}

	return nil
}

func toUseCaseRequest(eventPayload events2.PaymentInitPayload, id string) application.Request {
	return application.Request{
		UserID:        eventPayload.UserID,
		Amount:        eventPayload.Amount,
		CorrelationID: id,
	}
}

func NewSQSHandler(uc UseCase) *SQSHandler {
	return &SQSHandler{useCase: uc}
}
