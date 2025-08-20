package bootstrap

import (
	"github.com/payment-processor/internal/debit/application"
	"github.com/payment-processor/internal/debit/application/ports"
	"github.com/payment-processor/internal/debit/infra/handler"
)

func provideUseCase(repo ports.WalletRepository, bus ports.EventBusProcessor) *application.UseCaseHandler {
	return application.NewDebitBalanceUseCaseHandler(repo, bus)
}

func provideHandler(useCase *application.UseCaseHandler) *handler.SQSHandler {
	return handler.NewSQSHandler(useCase)
}
