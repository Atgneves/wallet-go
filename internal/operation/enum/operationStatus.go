package enum

type OperationStatus string

const (
	OperationStatusSuccess OperationStatus = "SUCCESS"
	OperationStatusError   OperationStatus = "ERROR"
)
