package bootstrap

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type LambdaHandler interface {
	Handle(ctx context.Context, sqsEvent events.SQSEvent) error
}

func BuildHandler() LambdaHandler {
	walletRepo := provideRepository()
	eventBus := provideEventBus()

	useCase := provideUseCase(walletRepo, eventBus)

	handler := provideHandler(useCase)

	return handler
}
