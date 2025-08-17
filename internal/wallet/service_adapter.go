package wallet

import (
	"context"
	"log"

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

func (sa *ServiceAdapter) DepositFromKafka(ctx context.Context, walletID uuid.UUID, amountInCents int64) error {
	request := WalletTransactionRequest{
		AmountInCents: amountInCents,
	}
	return sa.Deposit(ctx, walletID, request)
}

func (sa *ServiceAdapter) WithdrawFromKafka(ctx context.Context, walletID uuid.UUID, amountInCents int64) error {
	log.Printf("ServiceAdapter.WithdrawFromKafka - walletID: %s, amount: %d", walletID, amountInCents)
	request := WalletTransactionRequest{
		AmountInCents: amountInCents,
	}
	err := sa.Withdraw(ctx, walletID, request)
	if err != nil {
		log.Printf("ServiceAdapter.WithdrawFromKafka - ERROR: %v", err)
	} else {
		log.Printf("ServiceAdapter.WithdrawFromKafka - SUCCESS")
	}
	return err
}

func (sa *ServiceAdapter) TransferFromKafka(ctx context.Context, sourceID uuid.UUID, amountInCents int64, destinationID uuid.UUID) error {
	log.Printf("ServiceAdapter.TransferFromKafka - sourceID: %s, amount: %d, destinationID: %s", sourceID, amountInCents, destinationID)
	request := WalletTransactionTransferRequest{
		AmountInCents:       amountInCents,
		WalletDestinationID: destinationID,
	}
	err := sa.Transfer(ctx, sourceID, request)
	if err != nil {
		log.Printf("ServiceAdapter.TransferFromKafka - ERROR: %v", err)
	} else {
		log.Printf("ServiceAdapter.TransferFromKafka - SUCCESS")
	}
	return err
}
