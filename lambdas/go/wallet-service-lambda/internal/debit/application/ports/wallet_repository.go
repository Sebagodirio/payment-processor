package ports

import (
	"context"

	"github.com/payment-processor/internal/debit/domain"
)

type WalletRepository interface {
	Get(context.Context, domain.UserID) (domain.Wallet, error)
	Update(context.Context, domain.Wallet) error
}
