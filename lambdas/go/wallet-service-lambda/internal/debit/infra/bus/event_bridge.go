package bus

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/payment-processor/internal/debit/application/ports"
	"github.com/payment-processor/internal/debit/domain/events"
)

// ConsoleEventBus is a mock implementation of port EventBusProcessor.
// Simulates event publishing by printing in console
type ConsoleEventBus struct{}

func (b *ConsoleEventBus) Publish(ctx context.Context, req ports.BalanceDebitedRequest) error {
	event := events.BalanceDebitedEvent{
		Header: events.EventHeader{
			EventID:       uuid.NewString(),
			EventType:     string(req.EventName),
			Timestamp:     time.Now().UTC(),
			Version:       "1",
			CorrelationID: req.CorrelationID,
		},
		Payload: events.BalanceDebitedPayload{
			UserID:        req.UserID,
			AmountDebited: req.AmountDebited,
			AmountLeft:    req.AmountLeft,
		},
	}

	eventJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal event to JSON", "error", err)
		return err
	}

	slog.InfoContext(ctx, "--- EVENT PUBLISHED ---", "event", string(eventJSON))

	// real implementation would be like:
	// _, err := b.eventBridgeClient.PutEvents(...)
	// return err

	return nil
}

func NewConsoleEventBus() *ConsoleEventBus {
	return &ConsoleEventBus{}
}
