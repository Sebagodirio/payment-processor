package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/payment-processor/internal/debit/domain"
)

var (
	ErrWalletNotFound  = errors.New("wallet not found")
	ErrVersionMismatch = errors.New("optimistic lock failed: version mismatch")
)

type InMemoryWalletRepository struct {
	mu      sync.Mutex
	wallets map[domain.UserID]domain.Wallet
}

func (r *InMemoryWalletRepository) Get(_ context.Context, userID domain.UserID) (domain.Wallet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	wallet, ok := r.wallets[userID]
	if !ok {
		return domain.Wallet{}, ErrWalletNotFound
	}

	return wallet, nil
}

func (r *InMemoryWalletRepository) Update(_ context.Context, walletToUpdate domain.Wallet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentWallet, ok := r.wallets[walletToUpdate.UserID]
	if !ok {
		return ErrWalletNotFound
	}

	// Optimistic Blocking
	if currentWallet.Version != walletToUpdate.Version {
		return ErrVersionMismatch
	}

	walletToUpdate.Version++
	r.wallets[walletToUpdate.UserID] = walletToUpdate

	return nil
}

func NewInMemoryWalletRepository() *InMemoryWalletRepository {
	return &InMemoryWalletRepository{
		wallets: map[domain.UserID]domain.Wallet{
			"user-123": {
				UserID:  "user-123",
				Amount:  100.00,
				Version: 1, // Versión inicial
			},
			"user-456": {
				UserID:  "user-456",
				Amount:  50.00,
				Version: 1,
			},
		},
	}
}
