package ports

import (
	"context"

	"github.com/payment-processor/internal/debit/domain"
)

type BalanceDebitedRequest struct {
	UserID        domain.UserID
	AmountDebited domain.Amount
	AmountLeft    domain.Amount
	EventName     domain.Event
	CorrelationID string
}

type EventBusProcessor interface {
	Publish(context.Context, BalanceDebitedRequest) error
}
