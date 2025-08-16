package wallet

import (
	"context"

	"github.com/google/uuid"
)

// WalletService define a interface para operações de carteira
// Isso quebra a dependência circular entre kafka e wallet
type WalletService interface {
	Deposit(ctx context.Context, walletID uuid.UUID, request WalletTransactionRequest) (*Wallet, error)
	Withdraw(ctx context.Context, walletID uuid.UUID, request WalletTransactionRequest) (*Wallet, error)
	Transfer(ctx context.Context, sourceID uuid.UUID, request WalletTransactionTransferRequest) (*Wallet, error)
}
