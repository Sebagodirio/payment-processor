package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/payment-processor/internal/debit/application"
	"github.com/payment-processor/internal/debit/application/ports/mocks" // Importa mocks de los puertos
	"github.com/payment-processor/internal/debit/domain"
	"github.com/payment-processor/internal/debit/infra/repository" // Para el error de versión
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUseCaseHandler(t *testing.T) {
	t.Parallel()

	t.Run("should debit balance and publish event successfully", testUseCase_Success)
	t.Run("should return insufficient funds error when balance is too low", testUseCase_InsufficientFunds)
	t.Run("should return error when repository fails to get wallet", testUseCase_RepositoryGetError)
	t.Run("should return error when repository fails to update wallet", testUseCase_RepositoryUpdateError)
	t.Run("should succeed after one retry on version mismatch", testUseCase_OptimisticLockingRetrySuccess)
	t.Run("should fail after max retries on version mismatch", testUseCase_OptimisticLockingMaxRetries)
	t.Run("should return error when event bus fails to publish", testUseCase_EventBusError)
}

func testUseCase_Success(t *testing.T) {
	t.Parallel()

	// GIVEN
	repoMock := mocks.NewMockWalletRepository(t)
	busMock := mocks.NewMockEventBusProcessor(t)

	initialWallet := domain.Wallet{UserID: "user-123", Amount: 100, Version: 1}
	req := application.Request{UserID: "user-123", Amount: 30}

	repoMock.EXPECT().Get(mock.Anything, req.UserID).Return(initialWallet, nil).Once()
	repoMock.EXPECT().Update(mock.Anything, mock.MatchedBy(func(w domain.Wallet) bool {
		return w.UserID == req.UserID && w.Amount == 70 && w.Version == 1
	})).Return(nil).Once()
	busMock.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()

	useCase := application.NewDebitBalanceUseCaseHandler(repoMock, busMock)

	// WHEN
	err := useCase.Handle(context.Background(), req)

	// THEN
	assert.NoError(t, err)
}

func testUseCase_InsufficientFunds(t *testing.T) {
	t.Parallel()

	// GIVEN
	repoMock := mocks.NewMockWalletRepository(t)
	busMock := mocks.NewMockEventBusProcessor(t)

	initialWallet := domain.Wallet{UserID: "user-123", Amount: 20, Version: 1}
	req := application.Request{UserID: "user-123", Amount: 30}

	repoMock.EXPECT().Get(mock.Anything, req.UserID).Return(initialWallet, nil).Once()

	useCase := application.NewDebitBalanceUseCaseHandler(repoMock, busMock)

	// WHEN
	err := useCase.Handle(context.Background(), req)

	// THEN
	assert.Error(t, err)

	var domainErr *domain.Error
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "4001", domainErr.Code)

	repoMock.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	busMock.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func testUseCase_RepositoryGetError(t *testing.T) {
	t.Parallel()

	// GIVEN
	repoMock := mocks.NewMockWalletRepository(t)
	busMock := mocks.NewMockEventBusProcessor(t)
	req := application.Request{UserID: "user-123", Amount: 30}
	expectedError := errors.New("dynamo is down")

	repoMock.EXPECT().Get(mock.Anything, req.UserID).Return(domain.Wallet{}, expectedError).Once()

	useCase := application.NewDebitBalanceUseCaseHandler(repoMock, busMock)

	// WHEN
	err := useCase.Handle(context.Background(), req)

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get funds error")
}

func testUseCase_OptimisticLockingRetrySuccess(t *testing.T) {
	t.Parallel()

	// GIVEN
	repoMock := mocks.NewMockWalletRepository(t)
	busMock := mocks.NewMockEventBusProcessor(t)
	req := application.Request{UserID: "user-123", Amount: 30}

	walletV1 := domain.Wallet{UserID: "user-123", Amount: 100, Version: 1}
	walletV2 := domain.Wallet{UserID: "user-123", Amount: 80, Version: 2} // Otro proceso debitó 20

	repoMock.EXPECT().Get(mock.Anything, req.UserID).Return(walletV1, nil).Once()
	repoMock.EXPECT().Update(mock.Anything, mock.Anything).Return(repository.ErrVersionMismatch).Once()

	repoMock.EXPECT().Get(mock.Anything, req.UserID).Return(walletV2, nil).Once()
	repoMock.EXPECT().Update(mock.Anything, mock.MatchedBy(func(w domain.Wallet) bool {
		return w.Amount == 50
	})).Return(nil).Once()

	busMock.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()

	useCase := application.NewDebitBalanceUseCaseHandler(repoMock, busMock)

	// WHEN
	err := useCase.Handle(context.Background(), req)

	// THEN
	assert.NoError(t, err)
}

func testUseCase_RepositoryUpdateError(t *testing.T) {
	t.Parallel()

	// GIVEN
	repoMock := mocks.NewMockWalletRepository(t)
	busMock := mocks.NewMockEventBusProcessor(t)
	initialWallet := domain.Wallet{UserID: "user-123", Amount: 100, Version: 1}
	req := application.Request{UserID: "user-123", Amount: 30}
	expectedError := errors.New("unrecoverable db error")

	repoMock.EXPECT().Get(mock.Anything, req.UserID).Return(initialWallet, nil).Once()
	repoMock.EXPECT().Update(mock.Anything, mock.Anything).Return(expectedError).Once()

	useCase := application.NewDebitBalanceUseCaseHandler(repoMock, busMock)

	// WHEN
	err := useCase.Handle(context.Background(), req)

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "debit funds error")
}

func testUseCase_OptimisticLockingMaxRetries(t *testing.T) {
	t.Parallel()

	// GIVEN
	repoMock := mocks.NewMockWalletRepository(t)
	busMock := mocks.NewMockEventBusProcessor(t)
	req := application.Request{UserID: "user-123", Amount: 30}
	initialWallet := domain.Wallet{UserID: "user-123", Amount: 100, Version: 1}

	repoMock.EXPECT().Get(mock.Anything, req.UserID).Return(initialWallet, nil).Times(3) // Se llamará 3 veces (maxRetries)
	repoMock.EXPECT().Update(mock.Anything, mock.Anything).Return(repository.ErrVersionMismatch).Times(3)

	useCase := application.NewDebitBalanceUseCaseHandler(repoMock, busMock)

	// WHEN
	err := useCase.Handle(context.Background(), req)

	// THEN
	assert.Error(t, err)

	var domainErr *domain.Error
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "4002", domainErr.Code)
}

func testUseCase_EventBusError(t *testing.T) {
	t.Parallel()

	// GIVEN
	repoMock := mocks.NewMockWalletRepository(t)
	busMock := mocks.NewMockEventBusProcessor(t)
	initialWallet := domain.Wallet{UserID: "user-123", Amount: 100, Version: 1}
	req := application.Request{UserID: "user-123", Amount: 30}
	expectedError := errors.New("eventbridge is down")

	repoMock.EXPECT().Get(mock.Anything, req.UserID).Return(initialWallet, nil).Once()
	repoMock.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()

	busMock.EXPECT().Publish(mock.Anything, mock.Anything).Return(expectedError).Once()

	useCase := application.NewDebitBalanceUseCaseHandler(repoMock, busMock)

	// WHEN
	err := useCase.Handle(context.Background(), req)

	// THEN
	assert.Error(t, err)

	var domainErr *domain.Error
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "5003", domainErr.Code)
}
