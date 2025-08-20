package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/payment-processor/cmd/bootstrap"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	tp := bootstrap.InitTracing(ctx)

	if sdkTracerProvider, ok := tp.(*sdktrace.TracerProvider); ok {
		defer func() {
			if err := sdkTracerProvider.Shutdown(ctx); err != nil {
				slog.ErrorContext(ctx, "error shutting down tracer provider", "error", err)
			}
		}()
	}

	handler := bootstrap.BuildHandler()

	lambda.Start(otellambda.InstrumentHandler(handler.Handle))
}
