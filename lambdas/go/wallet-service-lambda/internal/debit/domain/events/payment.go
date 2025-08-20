package events

import (
	"time"

	"github.com/payment-processor/internal/debit/domain"
)

type EventHeader struct {
	EventID       string    `json:"event_id"`
	CorrelationID string    `json:"correlation_id"`
	EventType     string    `json:"event_type"`
	Timestamp     time.Time `json:"timestamp"`
	Version       string    `json:"version"`
}

type PaymentInitPayload struct {
	PaymentID     string        `json:"payment_id"`
	TransactionID string        `json:"transaction_id"`
	UserID        domain.UserID `json:"user_id"`
	Amount        domain.Amount `json:"amount"`
}

type PaymentInitEvent struct {
	Header  EventHeader        `json:"header"`
	Payload PaymentInitPayload `json:"payload"`
}
