package wallet

import (
	"time"

	"wallet-go/internal/operation"

	"github.com/google/uuid"
)

type Wallet struct {
	WalletID             uuid.UUID             `bson:"walletId" json:"walletId"`
	CustomerID           string                `bson:"customerId" json:"customerId"`
	CurrentAmountInCents int64                 `bson:"currentAmountInCents" json:"currentAmountInCents"`
	Operations           []operation.Operation `bson:"-" json:"operations,omitempty"` // ← NÃO salvar no MongoDB (bson:"-")
	Active               bool                  `bson:"active" json:"active"`
	Blocked              bool                  `bson:"blocked" json:"blocked"`
	CreatedAt            time.Time             `bson:"createdAt" json:"createdAt"`
	UpdatedAt            time.Time             `bson:"updatedAt" json:"updatedAt"`
	BlockedAt            *time.Time            `bson:"blockedAt,omitempty" json:"blockedAt,omitempty"`
	UnblockedAt          *time.Time            `bson:"unblockedAt,omitempty" json:"unblockedAt,omitempty"`
}

type WalletRequest struct {
	CustomerID string `json:"customerId" validate:"required" binding:"required"`
}

type WalletPatch struct {
	Active  *bool `json:"active,omitempty"`
	Blocked *bool `json:"blocked,omitempty"`
}

type WalletTransactionRequest struct {
	AmountInCents int64 `json:"amountInCents" validate:"required,gt=0" binding:"required,gt=0"`
}

type WalletTransactionTransferRequest struct {
	AmountInCents       int64     `json:"amountInCents" validate:"required,gt=0" binding:"required,gt=0"`
	WalletDestinationID uuid.UUID `json:"walletDestinationId" validate:"required" binding:"required"`
}

type WalletKafkaTransactionMessage struct {
	WalletID      uuid.UUID `json:"walletId"`
	AmountInCents int64     `json:"amountInCents"`
}

type WalletKafkaTransactionTransferMessage struct {
	WalletID            uuid.UUID `json:"walletId"`
	AmountInCents       int64     `json:"amountInCents"`
	WalletDestinationID uuid.UUID `json:"walletDestinationId"`
}

type WalletResponse struct {
	Id                   uuid.UUID             `json:"id"`
	CustomerID           string                `json:"customerId"`
	CurrentAmountInCents int64                 `json:"currentAmountInCents"`
	Operations           []operation.Operation `json:"operations,omitempty"`
	Active               bool                  `json:"active"`
	Blocked              bool                  `json:"blocked"`
	CreatedAt            time.Time             `json:"createdAt"`
	UpdatedAt            time.Time             `json:"updatedAt"`
	BlockedAt            *time.Time            `json:"blockedAt,omitempty"`
	UnblockedAt          *time.Time            `json:"unblockedAt,omitempty"`
}

// Wallet methods
func (w *Wallet) IsActive() bool {
	return w.Active
}

func (w *Wallet) IsBlocked() bool {
	return w.Blocked
}

func (w *Wallet) WithActive(status bool) {
	w.Active = status
	w.UpdatedAt = time.Now()
}

func (w *Wallet) block() {
	w.Blocked = true
	now := time.Now()
	w.BlockedAt = &now
	w.UpdatedAt = now
}

func (w *Wallet) unblock() {
	w.Blocked = false
	now := time.Now()
	w.UnblockedAt = &now
	w.UpdatedAt = now
}

func (w *Wallet) ChangeBlock(isBlocked bool) {
	if isBlocked {
		w.block()
	} else {
		w.unblock()
	}
}

func (w *Wallet) HasBalanceToDebit(amountInCents int64) bool {
	newCurrentAmount := w.CurrentAmountInCents - amountInCents
	return newCurrentAmount >= 0
}

func (w *Wallet) IncrementCurrentAmountInCents(amountInCents int64) {
	w.CurrentAmountInCents += amountInCents
}

func (w *Wallet) DecreaseCurrentAmountInCents(amountInCents int64) {
	newCurrentAmount := w.CurrentAmountInCents - amountInCents
	if newCurrentAmount < 0 {
		w.CurrentAmountInCents = 0
	} else {
		w.CurrentAmountInCents = newCurrentAmount
	}
}
