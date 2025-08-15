package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int    `json:"status"`
	Type    string `json:"error"` // Mudei de "Error" para "Type"
	Message string `json:"message"`
}

func (e *AppError) Error() string { // Método necessário para interface error
	return fmt.Sprintf("Code: %d, Error: %s, Message: %s", e.Code, e.Type, e.Message)
}

func (e *AppError) String() string { // Pode manter este também se quiser
	return e.Error()
}

// Wallet errors
func WalletNotFound() *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Type:    "Not Found",
		Message: "Wallet not found!",
	}
}

func WalletInactive() *AppError {
	return &AppError{
		Code:    http.StatusUnprocessableEntity,
		Type:    "Unprocessable Entity",
		Message: "Wallet is inactive!",
	}
}

func WalletBlocked() *AppError {
	return &AppError{
		Code:    http.StatusUnprocessableEntity,
		Type:    "Unprocessable Entity",
		Message: "Wallet is blocked!",
	}
}

func WalletInsufficientFunds() *AppError {
	return &AppError{
		Code:    http.StatusUnprocessableEntity,
		Type:    "Unprocessable Entity",
		Message: "Insufficient wallet funds!",
	}
}

func SameWalletTransferNotAllowed() *AppError {
	return &AppError{
		Code:    http.StatusUnprocessableEntity,
		Type:    "Unprocessable Entity",
		Message: "The source and destination wallets must be different!",
	}
}

func WalletBadRequest(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Type:    "Bad Request",
		Message: message,
	}
}

// Operation errors
func OperationNotFound() *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Type:    "Not Found",
		Message: "Operation not found!",
	}
}

// Generic errors
func InternalServerError(message string) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Type:    "Internal Server Error",
		Message: message,
	}
}

func BadRequest(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Type:    "Bad Request",
		Message: message,
	}
}
