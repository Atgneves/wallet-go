package enum

type OperationType string

const (
	OperationTypeCreated         OperationType = "CREATED"
	OperationTypeDeposit         OperationType = "DEPOSIT"
	OperationTypeWithdraw        OperationType = "WITHDRAW"
	OperationTypeTransfer        OperationType = "TRANSFER"
	OperationTypeReceiveTransfer OperationType = "RECEIVE_TRANSFER"
)
