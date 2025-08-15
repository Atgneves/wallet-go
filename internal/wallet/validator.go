package wallet

import (
	"fmt"
	"github.com/Atgneves/wallet-go/internal/shared/errors"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) EnsureValidForOperation(wallet *Wallet, context string) (*Wallet, error) {
	if wallet == nil {
		return nil, errors.WalletNotFound()
	}

	if !wallet.IsActive() {
		return nil, &errors.AppError{
			Code:    422,
			Type:    "Unprocessable Entity",
			Message: fmt.Sprintf("Cannot process transaction. %s is inactive!", context),
		}
	}

	if wallet.IsBlocked() {
		return nil, &errors.AppError{
			Code:    422,
			Type:    "Unprocessable Entity",
			Message: fmt.Sprintf("Cannot process transaction. %s is blocked!", context),
		}
	}

	return wallet, nil
}

func (v *Validator) HasBalanceToDebit(wallet *Wallet, context string, amountInCents int64) error {
	if !wallet.HasBalanceToDebit(amountInCents) {
		return &errors.AppError{
			Code:    422,
			Type:    "Unprocessable Entity",
			Message: fmt.Sprintf("Cannot process transaction. Insufficient balance %s!", context),
		}
	}
	return nil
}

func (v *Validator) ValidateForDebitOperation(wallet *Wallet, context string, amountInCents int64) error {
	if _, err := v.EnsureValidForOperation(wallet, context); err != nil {
		return err
	}
	return v.HasBalanceToDebit(wallet, context, amountInCents)
}
