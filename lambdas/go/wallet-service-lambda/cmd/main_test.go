package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/payment-processor/bootstrap"
	_events "github.com/payment-processor/internal/debit/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLambdaHandler_EndToEnd es el único archivo en español y para tener una prueba del sistema mockeado
// En un entorno real, este archivo no existiría y los test e2e se harían en un ambiente controlado no productivo
func TestLambdaHandler_EndToEnd(t *testing.T) {
	// --- 1. Preparación ---

	// Construimos nuestro handler con todas sus dependencias (mocks)
	// exactamente como lo haría el main.go real.
	handler := bootstrap.BuildHandler()

	// Creamos el evento de entrada que simula lo que llegaría en un mensaje de SQS.
	inputEvent := _events.PaymentInitEvent{
		Header: _events.EventHeader{
			CorrelationID: "test-correlation-id-123",
		},
		Payload: _events.PaymentInitPayload{
			UserID: "user-123", // Este usuario empieza con 100.00
			Amount: 25.50,
		},
	}

	// Convertimos el evento a JSON, que es como viaja en el cuerpo del mensaje.
	eventBody, err := json.Marshal(inputEvent)
	require.NoError(t, err)

	// Creamos el evento de SQS completo, que es lo que recibe la función Handle.
	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "test-message-id",
				Body:      string(eventBody),
			},
		},
	}

	// --- 2. Actuación  ---

	// Invocamos el handler con nuestro evento simulado.
	// Esto ejecutará todo el flujo: handler -> use case -> repository -> bus.
	err = handler.Handle(context.Background(), sqsEvent)

	// --- 3. Aserción ---

	// Verificamos que el handler no haya devuelto ningún error.
	assert.NoError(t, err)
}
