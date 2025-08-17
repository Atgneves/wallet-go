package operation

import (
	"context"
	"time"
	"wallet-go/internal/operation/enum"

	"wallet-go/internal/shared/errors"

	"github.com/google/uuid"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) GetByWalletID(ctx context.Context, walletID uuid.UUID) ([]Operation, error) {
	operationPointers, err := s.store.FindByWalletID(ctx, walletID)
	if err != nil {
		return nil, errors.InternalServerError("Failed to get operations")
	}

	// Converter []*Operation para []Operation
	operations := make([]Operation, len(operationPointers))
	for i, op := range operationPointers {
		operations[i] = *op
	}

	return operations, nil
}

func (s *Service) GetByID(ctx context.Context, operationID uuid.UUID) (*Operation, error) {
	operation, err := s.store.FindByID(ctx, operationID)
	if err != nil {
		return nil, errors.InternalServerError("Failed to get operation")
	}

	if operation == nil {
		return nil, errors.OperationNotFound()
	}

	return operation, nil
}

func (s *Service) List(ctx context.Context, request OperationFilterRequest) ([]*Operation, error) {
	from, err := time.Parse("2006-01-02", request.From)
	if err != nil {
		return nil, errors.BadRequest("Invalid from date format")
	}

	to, err := time.Parse("2006-01-02", request.To)
	if err != nil {
		return nil, errors.BadRequest("Invalid to date format")
	}

	// Add one day to include the full "to" date
	to = to.Add(24 * time.Hour)

	operations, err := s.store.FindByWalletIDAndDateRange(ctx, request.WalletID, from, to)
	if err != nil {
		return nil, errors.InternalServerError("Failed to list operations")
	}

	return operations, nil
}

func (s *Service) GetDailySummary(ctx context.Context, walletID uuid.UUID, date time.Time) (*OperationDailySummaryResponse, error) {
	if err := s.validateDailySummaryParams(walletID, date); err != nil {
		return nil, err
	}

	summaryData, err := s.buildDailySummaryData(ctx, walletID, date)
	if err != nil {
		return nil, err
	}

	return s.buildSummaryResponse(summaryData, nil), nil
}

func (s *Service) GetDailySummaryDetails(ctx context.Context, walletID uuid.UUID, date time.Time) (*OperationDailySummaryResponse, error) {
	if err := s.validateDailySummaryParams(walletID, date); err != nil {
		return nil, err
	}

	summaryData, err := s.buildDailySummaryData(ctx, walletID, date)
	if err != nil {
		return nil, err
	}

	operationItems := s.mapOperationsToItems(summaryData.Operations)
	return s.buildSummaryResponse(summaryData, operationItems), nil
}

func (s *Service) validateDailySummaryParams(walletID uuid.UUID, date time.Time) error {
	if walletID == uuid.Nil {
		return errors.BadRequest("Wallet ID cannot be null!")
	}

	if date.IsZero() {
		return errors.BadRequest("Date cannot be null")
	}

	today := time.Now().Truncate(24 * time.Hour)
	if date.After(today) {
		return errors.BadRequest("Summary request cannot be for future dates")
	}

	return nil
}

type dailySummaryData struct {
	WalletID         uuid.UUID
	Date             time.Time
	Operations       []*Operation
	WalletBalanceDay int64
}

func (s *Service) buildDailySummaryData(ctx context.Context, walletID uuid.UUID, date time.Time) (*dailySummaryData, error) {
	operations, err := s.store.FindByWalletIDAndDate(ctx, walletID, date)
	if err != nil {
		return nil, errors.InternalServerError("Failed to get operations")
	}

	if len(operations) == 0 {
		return &dailySummaryData{
			WalletID:         walletID,
			Date:             date,
			Operations:       []*Operation{},
			WalletBalanceDay: 0,
		}, nil
	}

	// Calculate balance considering only successful operations of allowed types
	var walletBalanceDay int64
	allowedTypes := map[enum.OperationType]bool{
		enum.OperationTypeDeposit:         true,
		enum.OperationTypeWithdraw:        true,
		enum.OperationTypeTransfer:        true,
		enum.OperationTypeReceiveTransfer: true,
	}

	for _, op := range operations {
		if op.Status == enum.OperationStatusSuccess && allowedTypes[op.Type] {
			walletBalanceDay += op.AmountInCents
		}
	}

	return &dailySummaryData{
		WalletID:         walletID,
		Date:             date,
		Operations:       operations,
		WalletBalanceDay: walletBalanceDay,
	}, nil
}

func (s *Service) mapOperationsToItems(operations []*Operation) []OperationItem {
	if len(operations) == 0 {
		return []OperationItem{}
	}

	items := make([]OperationItem, len(operations))
	for i, op := range operations {
		items[i] = OperationItem{
			Type:          op.Type,
			Status:        op.Status,
			AmountInCents: op.AmountInCents,
			CreatedAt:     op.CreatedAt,
		}
	}

	return items
}

func (s *Service) buildSummaryResponse(summaryData *dailySummaryData, operationItems []OperationItem) *OperationDailySummaryResponse {
	return &OperationDailySummaryResponse{
		WalletID:          summaryData.WalletID,
		DateBalanceWallet: summaryData.Date.Format("2006-01-02"),
		WalletBalanceDay:  summaryData.WalletBalanceDay,
		Operations:        operationItems,
	}
}
