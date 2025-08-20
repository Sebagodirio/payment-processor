package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/payment-processor/internal/debit/application"
	"github.com/payment-processor/internal/debit/domain"
	_events "github.com/payment-processor/internal/debit/domain/events"
	"github.com/payment-processor/internal/debit/infra/handler"
	"github.com/payment-processor/internal/debit/infra/handler/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSQSHandler(t *testing.T) {
	t.Parallel()

	t.Run("should process message successfully", testHandlerSuccessfully)
	t.Run("should return error when message body is invalid json", testHandlerUnmarshalError)
	t.Run("should not return error when event validation fails", testHandlerValidationError)
	t.Run("should return error when use case fails", testHandlerUseCaseError)
}

func testHandlerSuccessfully(t *testing.T) {
	t.Parallel()

	// GIVEN
	useCaseMock := mocks.NewMockUseCase(t)

	useCaseRequest := application.Request{
		UserID:        "user-123",
		Amount:        50.5,
		CorrelationID: "corr-id-abc",
	}

	sqsEvent := createSQSEvent(t, useCaseRequest.UserID, useCaseRequest.Amount, useCaseRequest.CorrelationID)

	useCaseMock.EXPECT().Handle(mock.Anything, useCaseRequest).Return(nil).Once()

	h := handler.NewSQSHandler(useCaseMock)

	// WHEN
	err := h.Handle(context.Background(), sqsEvent)

	// THEN
	assert.NoError(t, err)
}

func testHandlerUnmarshalError(t *testing.T) {
	t.Parallel()

	// GIVEN
	useCaseMock := mocks.NewMockUseCase(t)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{{Body: "this is not json"}},
	}
	h := handler.NewSQSHandler(useCaseMock)

	// WHEN
	err := h.Handle(context.Background(), sqsEvent)

	// THEN
	assert.Error(t, err)
}

func testHandlerValidationError(t *testing.T) {
	t.Parallel()

	// GIVEN
	useCaseMock := mocks.NewMockUseCase(t)
	sqsEvent := createSQSEvent(t, "", 50.5, "corr-id-abc")
	h := handler.NewSQSHandler(useCaseMock)

	// WHEN
	err := h.Handle(context.Background(), sqsEvent)

	// THEN
	assert.NoError(t, err)
	useCaseMock.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything)
}

func testHandlerUseCaseError(t *testing.T) {
	t.Parallel()

	// GIVEN
	useCaseMock := mocks.NewMockUseCase(t)
	expectedError := errors.New("something went wrong in the use case")
	useCaseRequest := application.Request{
		UserID:        "user-123",
		Amount:        50.5,
		CorrelationID: "corr-id-abc",
	}
	sqsEvent := createSQSEvent(t, useCaseRequest.UserID, useCaseRequest.Amount, useCaseRequest.CorrelationID)

	useCaseMock.EXPECT().Handle(mock.Anything, useCaseRequest).Return(expectedError).Once()

	h := handler.NewSQSHandler(useCaseMock)

	// WHEN
	err := h.Handle(context.Background(), sqsEvent)

	// THEN

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// --- Helper Functions ---

func createSQSEvent(t *testing.T, userID domain.UserID, amount domain.Amount, corrID string) events.SQSEvent {
	t.Helper()

	eventPayload := _events.PaymentInitPayload{
		UserID: userID,
		Amount: amount,
	}
	event := _events.PaymentInitEvent{
		Header:  _events.EventHeader{CorrelationID: corrID},
		Payload: eventPayload,
	}

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	return events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "test-message-id",
				Body:      string(body),
			},
		},
	}
}
