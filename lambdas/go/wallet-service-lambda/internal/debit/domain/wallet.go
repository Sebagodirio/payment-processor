package domain

var (
	BalanceDebitedEventName Event = "BalanceDebited"
)

type (
	Event  string
	UserID string
	Amount float64
)

type Wallet struct {
	UserID  UserID
	Amount  Amount
	Version int
}

func (w *Wallet) CanWithdraw(amountToWithdraw Amount) bool {
	return w.Amount >= amountToWithdraw
}

func (w *Wallet) Debit(amountToDebit Amount) error {
	if !w.CanWithdraw(amountToDebit) {
		return NewInsufficientFundsError(string(w.UserID), float64(w.Amount), float64(amountToDebit))
	}
	w.Amount -= amountToDebit
	return nil
}
