package wallet

import (
	"context"
	"fmt"
	"time"
	"wallet-go/internal/operation/enum"

	"wallet-go/internal/operation"
	"wallet-go/internal/shared/errors"
	"wallet-go/internal/shared/utils"

	"github.com/google/uuid"
)

type Service struct {
	store          *Store
	operationStore *operation.Store
	validator      *Validator
	lockManager    *utils.WalletLockManager
}

func NewService(store *Store, operationStore *operation.Store, validator *Validator, lockManager *utils.WalletLockManager) *Service {
	return &Service{
		store:          store,
		operationStore: operationStore,
		validator:      validator,
		lockManager:    lockManager,
	}
}

func (s *Service) Create(ctx context.Context, request WalletRequest) (*Wallet, error) {
	// Check if wallet already exists for customer
	existingWallet, err := s.store.FindByCustomerID(ctx, request.CustomerID)
	if err != nil {
		return nil, errors.InternalServerError("Failed to check existing wallet")
	}

	if existingWallet != nil {
		return existingWallet, nil
	}

	// Create new wallet
	walletID := uuid.New()
	now := time.Now()

	wallet := &Wallet{
		WalletID:             walletID,
		CustomerID:           request.CustomerID,
		CurrentAmountInCents: 0,
		Active:               true,
		Blocked:              false,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := s.store.Create(ctx, wallet); err != nil {
		return nil, errors.InternalServerError("Failed to create wallet")
	}

	// Create creation operation
	createOperation := &operation.Operation{
		OperationID:   uuid.New(),
		WalletID:      walletID,
		Type:          enum.OperationTypeCreated,
		Status:        enum.OperationStatusSuccess,
		AmountInCents: 0,
		Reason:        "Created wallet success!",
		CreatedAt:     now,
	}

	if err := s.operationStore.Create(ctx, createOperation); err != nil {
		return nil, errors.InternalServerError("Failed to create operation")
	}

	return wallet, nil
}

func (s *Service) GetByID(ctx context.Context, walletID uuid.UUID) (*Wallet, error) {
	wallet, err := s.store.FindByID(ctx, walletID)
	if err != nil {
		return nil, errors.InternalServerError("Failed to get wallet")
	}

	if wallet == nil {
		return nil, errors.WalletNotFound()
	}

	return wallet, nil
}

func (s *Service) List(ctx context.Context) ([]*Wallet, error) {
	fmt.Printf("DEBUG: WalletService.List() - iniciando busca no store...\n")

	wallets, err := s.store.FindAll(ctx)
	if err != nil {
		fmt.Printf("DEBUG: WalletService.List() - erro no store.FindAll(): %v\n", err)
		return nil, errors.InternalServerError("Failed to list wallets")
	}

	fmt.Printf("DEBUG: WalletService.List() - store retornou %d wallets\n", len(wallets))
	return wallets, nil
}

func (s *Service) Patch(ctx context.Context, walletID uuid.UUID, patch WalletPatch) (*Wallet, error) {
	wallet, err := s.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if patch.Active != nil {
		wallet.WithActive(*patch.Active)
	}

	if patch.Blocked != nil {
		wallet.ChangeBlock(*patch.Blocked)
	}

	if err := s.store.Update(ctx, wallet); err != nil {
		return nil, errors.InternalServerError("Failed to update wallet")
	}

	return wallet, nil
}

func (s *Service) Deposit(ctx context.Context, walletID uuid.UUID, request WalletTransactionRequest) (*Wallet, error) {
	s.lockManager.LockWallet(walletID)
	defer s.lockManager.UnlockWallet(walletID)

	wallet, err := s.getWalletOrThrow(ctx, walletID)
	if err != nil {
		return nil, err
	}

	validatedWallet, err := s.validator.EnsureValidForOperation(wallet, "Wallet")
	if err != nil {
		s.handleErrorOperation(ctx, wallet, enum.OperationTypeDeposit, request.AmountInCents, err.(*errors.AppError).Message)
		return nil, err
	}

	return s.executeDeposit(ctx, validatedWallet, request)
}

func (s *Service) Withdraw(ctx context.Context, walletID uuid.UUID, request WalletTransactionRequest) (*Wallet, error) {
	s.lockManager.LockWallet(walletID)
	defer s.lockManager.UnlockWallet(walletID)

	wallet, err := s.getWalletOrThrow(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if err := s.validator.ValidateForDebitOperation(wallet, "Source wallet", request.AmountInCents); err != nil {
		s.handleErrorOperation(ctx, wallet, enum.OperationTypeWithdraw, -request.AmountInCents, err.(*errors.AppError).Message)
		return nil, err
	}

	return s.executeWithdraw(ctx, wallet, request)
}

func (s *Service) Transfer(ctx context.Context, sourceID uuid.UUID, request WalletTransactionTransferRequest) (*Wallet, error) {
	// Sort IDs to prevent deadlocks
	var firstID, secondID uuid.UUID
	if sourceID.String() < request.WalletDestinationID.String() {
		firstID, secondID = sourceID, request.WalletDestinationID
	} else {
		firstID, secondID = request.WalletDestinationID, sourceID
	}

	s.lockManager.LockWallet(firstID)
	defer s.lockManager.UnlockWallet(firstID)
	s.lockManager.LockWallet(secondID)
	defer s.lockManager.UnlockWallet(secondID)

	sourceWallet, err := s.getWalletOrThrow(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	destinationWallet, err := s.getWalletOrThrow(ctx, request.WalletDestinationID)
	if err != nil {
		return nil, err
	}

	if sourceWallet.WalletID == destinationWallet.WalletID {
		s.handleErrorOperation(ctx, sourceWallet, enum.OperationTypeTransfer, -request.AmountInCents,
			"Cannot process transaction. The source and destination wallets must be different!")
		return nil, errors.SameWalletTransferNotAllowed()
	}

	if err := s.validator.ValidateForDebitOperation(sourceWallet, "Source wallet", request.AmountInCents); err != nil {
		s.handleErrorOperation(ctx, sourceWallet, enum.OperationTypeTransfer, -request.AmountInCents, err.(*errors.AppError).Message)
		return nil, err
	}

	if _, err := s.validator.EnsureValidForOperation(destinationWallet, "Destination wallet"); err != nil {
		s.handleErrorOperation(ctx, sourceWallet, enum.OperationTypeTransfer, -request.AmountInCents, err.(*errors.AppError).Message)
		return nil, err
	}

	return s.executeTransfer(ctx, sourceWallet, destinationWallet, request)
}

func (s *Service) getWalletOrThrow(ctx context.Context, walletID uuid.UUID) (*Wallet, error) {
	wallet, err := s.store.FindByID(ctx, walletID)
	if err != nil {
		return nil, errors.InternalServerError("Failed to get wallet")
	}

	if wallet == nil {
		return nil, errors.WalletNotFound()
	}

	return wallet, nil
}

func (s *Service) executeDeposit(ctx context.Context, wallet *Wallet, request WalletTransactionRequest) (*Wallet, error) {
	wallet.IncrementCurrentAmountInCents(request.AmountInCents)

	if err := s.store.Update(ctx, wallet); err != nil {
		return nil, errors.InternalServerError("Failed to update wallet")
	}

	op := &operation.Operation{
		OperationID:   uuid.New(),
		WalletID:      wallet.WalletID,
		Type:          enum.OperationTypeDeposit,
		Status:        enum.OperationStatusSuccess,
		AmountInCents: request.AmountInCents,
		Reason:        "Deposit success!",
		CreatedAt:     time.Now(),
	}

	if err := s.operationStore.Create(ctx, op); err != nil {
		return nil, errors.InternalServerError("Failed to create operation")
	}

	return wallet, nil
}

func (s *Service) executeWithdraw(ctx context.Context, wallet *Wallet, request WalletTransactionRequest) (*Wallet, error) {
	wallet.DecreaseCurrentAmountInCents(request.AmountInCents)

	if err := s.store.Update(ctx, wallet); err != nil {
		return nil, errors.InternalServerError("Failed to update wallet")
	}

	op := &operation.Operation{
		OperationID:   uuid.New(),
		WalletID:      wallet.WalletID,
		Type:          enum.OperationTypeWithdraw,
		Status:        enum.OperationStatusSuccess,
		AmountInCents: -request.AmountInCents,
		Reason:        "Withdraw success!",
		CreatedAt:     time.Now(),
	}

	if err := s.operationStore.Create(ctx, op); err != nil {
		return nil, errors.InternalServerError("Failed to create operation")
	}

	return wallet, nil
}

func (s *Service) executeTransfer(ctx context.Context, sourceWallet, destinationWallet *Wallet, request WalletTransactionTransferRequest) (*Wallet, error) {
	operationIDSource := uuid.New()
	operationIDDestination := uuid.New()

	sourceWallet.DecreaseCurrentAmountInCents(request.AmountInCents)
	destinationWallet.IncrementCurrentAmountInCents(request.AmountInCents)

	if err := s.store.Update(ctx, sourceWallet); err != nil {
		return nil, errors.InternalServerError("Failed to update source wallet")
	}

	if err := s.store.Update(ctx, destinationWallet); err != nil {
		return nil, errors.InternalServerError("Failed to update destination wallet")
	}

	transferOp := &operation.Operation{
		OperationID:            operationIDSource,
		WalletID:               sourceWallet.WalletID,
		Type:                   enum.OperationTypeTransfer,
		Status:                 enum.OperationStatusSuccess,
		AmountInCents:          -request.AmountInCents,
		WalletTransactionID:    &destinationWallet.WalletID,
		OperationTransactionID: &operationIDDestination,
		Reason:                 "Transfer success!",
		CreatedAt:              time.Now(),
	}

	receiveOp := &operation.Operation{
		OperationID:            operationIDDestination,
		WalletID:               destinationWallet.WalletID,
		Type:                   enum.OperationTypeReceiveTransfer,
		Status:                 enum.OperationStatusSuccess,
		AmountInCents:          request.AmountInCents,
		WalletTransactionID:    &sourceWallet.WalletID,
		OperationTransactionID: &operationIDSource,
		Reason:                 "Transfer received success!",
		CreatedAt:              time.Now(),
	}

	if err := s.operationStore.Create(ctx, transferOp); err != nil {
		return nil, errors.InternalServerError("Failed to create transfer operation")
	}

	if err := s.operationStore.Create(ctx, receiveOp); err != nil {
		return nil, errors.InternalServerError("Failed to create receive operation")
	}

	return sourceWallet, nil
}

func (s *Service) handleErrorOperation(ctx context.Context, wallet *Wallet, opType enum.OperationType, amountInCents int64, message string) {
	if wallet == nil {
		return
	}

	errorOp := &operation.Operation{
		OperationID:   uuid.New(),
		WalletID:      wallet.WalletID,
		Type:          opType,
		Status:        enum.OperationStatusError,
		AmountInCents: amountInCents,
		Reason:        message,
		CreatedAt:     time.Now(),
	}

	s.operationStore.Create(ctx, errorOp)
	s.store.Update(ctx, wallet)
}
