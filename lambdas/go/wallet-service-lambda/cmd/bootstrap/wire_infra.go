package bootstrap

import (
	"github.com/payment-processor/internal/debit/infra/bus"
	"github.com/payment-processor/internal/debit/infra/repository"
)

func provideRepository() *repository.InMemoryWalletRepository {
	// in a real case, we would instance the real client here
	return repository.NewInMemoryWalletRepository()
}

func provideEventBus() *bus.ConsoleEventBus {
	// in a real case, we would instance the real client here
	return bus.NewConsoleEventBus()
}
