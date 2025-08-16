package operation

import (
	"time"

	"wallet-go/internal/operation/enum"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Operation struct {
	ID                     primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	OperationID            uuid.UUID            `bson:"operationId" json:"operationId"`
	WalletID               uuid.UUID            `bson:"walletId" json:"walletId"`
	Type                   enum.OperationType   `bson:"type" json:"type"`
	Status                 enum.OperationStatus `bson:"status" json:"status"`
	AmountInCents          int64                `bson:"amountInCents" json:"amountInCents"`
	WalletTransactionID    *uuid.UUID           `bson:"walletTransactionId,omitempty" json:"walletTransactionId,omitempty"`
	OperationTransactionID *uuid.UUID           `bson:"operationTransactionId,omitempty" json:"operationTransactionId,omitempty"`
	Reason                 string               `bson:"reason" json:"reason"`
	CreatedAt              time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt              *time.Time           `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

type OperationResponse struct {
	ID                     uuid.UUID            `json:"id"`
	WalletID               uuid.UUID            `json:"walletId"`
	Type                   enum.OperationType   `json:"type"`
	Status                 enum.OperationStatus `json:"status"`
	AmountInCents          int64                `json:"amountInCents"`
	WalletTransactionID    *uuid.UUID           `json:"walletTransactionId,omitempty"`
	OperationTransactionID *uuid.UUID           `json:"operationTransactionId,omitempty"`
	Reason                 string               `json:"reason"`
	CreatedAt              time.Time            `json:"createdAt"`
	UpdatedAt              *time.Time           `json:"updatedAt,omitempty"`
}

type OperationFilterRequest struct {
	WalletID uuid.UUID `form:"walletId" binding:"required"`
	From     string    `form:"from" binding:"required"`
	To       string    `form:"to" binding:"required"`
}

type OperationDailySummaryResponse struct {
	WalletID          uuid.UUID       `json:"walletId"`
	DateBalanceWallet string          `json:"dateBalanceWallet"`
	WalletBalanceDay  int64           `json:"walletBalanceDay"`
	Operations        []OperationItem `json:"operations,omitempty"`
}

type OperationItem struct {
	Type          enum.OperationType   `json:"type"`
	Status        enum.OperationStatus `json:"status"`
	AmountInCents int64                `json:"amountInCents"`
	CreatedAt     time.Time            `json:"createdAt"`
}
