package events

import "github.com/payment-processor/internal/debit/domain"

type BalanceDebitedPayload struct {
	UserID        domain.UserID `json:"userId"`
	AmountDebited domain.Amount `json:"amountDebited"`
	AmountLeft    domain.Amount `json:"amountLeft"`
}

type BalanceDebitedEvent struct {
	Header  EventHeader           `json:"header"`
	Payload BalanceDebitedPayload `json:"payload"`
}
