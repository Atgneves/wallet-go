package wallet

import (
	"context"

	"github.com/google/uuid"
)

// ServiceAdapter adapta o Service para implementar a interface do kafka
// Isso permite usar o service no kafka consumer sem dependência circular
type ServiceAdapter struct {
	service *Service
}

func NewServiceAdapter(service *Service) *ServiceAdapter {
	return &ServiceAdapter{
		service: service,
	}
}

// Implementa a interface do kafka.WalletService
func (sa *ServiceAdapter) Deposit(ctx context.Context, walletID uuid.UUID, request WalletTransactionRequest) error {
	_, err := sa.service.Deposit(ctx, walletID, request)
	return err
}

func (sa *ServiceAdapter) Withdraw(ctx context.Context, walletID uuid.UUID, request WalletTransactionRequest) error {
	_, err := sa.service.Withdraw(ctx, walletID, request)
	return err
}

func (sa *ServiceAdapter) Transfer(ctx context.Context, sourceID uuid.UUID, request WalletTransactionTransferRequest) error {
	_, err := sa.service.Transfer(ctx, sourceID, request)
	return err
}

// Método auxiliar para converter de kafka types para wallet types
func (sa *ServiceAdapter) DepositFromKafka(ctx context.Context, walletID uuid.UUID, amountInCents int64) error {
	request := WalletTransactionRequest{
		AmountInCents: amountInCents,
	}
	return sa.Deposit(ctx, walletID, request)
}

func (sa *ServiceAdapter) WithdrawFromKafka(ctx context.Context, walletID uuid.UUID, amountInCents int64) error {
	request := WalletTransactionRequest{
		AmountInCents: amountInCents,
	}
	return sa.Withdraw(ctx, walletID, request)
}

func (sa *ServiceAdapter) TransferFromKafka(ctx context.Context, sourceID uuid.UUID, amountInCents int64, destinationID uuid.UUID) error {
	request := WalletTransactionTransferRequest{
		AmountInCents:       amountInCents,
		WalletDestinationID: destinationID,
	}
	return sa.Transfer(ctx, sourceID, request)
}
