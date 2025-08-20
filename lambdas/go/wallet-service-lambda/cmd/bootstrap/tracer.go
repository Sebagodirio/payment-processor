package bootstrap

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func newTracerProvider() (*sdktrace.TracerProvider, error) {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithIDGenerator(xray.NewIDGenerator()),
	)
	return tp, nil
}

func InitTracing(ctx context.Context) trace.TracerProvider {
	tp, err := newTracerProvider()
	if err != nil {
		slog.ErrorContext(ctx, "failed to initialize tracer provider", "error", err)
		return noop.NewTracerProvider()
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	slog.InfoContext(ctx, "X-Ray tracer provider initialized")
	return tp
}
