package main

import (
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/payment-processor/bootstrap"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	handler := bootstrap.BuildHandler()

	lambda.Start(handler.Handle)
}
