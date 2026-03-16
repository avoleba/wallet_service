package model

import (
	"github.com/google/uuid"
	"time"
)

type OperationType string

const (
	OperationDeposit  OperationType = "DEPOSIT"
	OperationWithdraw OperationType = "WITHDRAW"
)

type Wallet struct {
	ID        uuid.UUID `json:"id"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WalletOperationRequest struct {
	WalletID      uuid.UUID     `json:"walletId"`
	OperationType OperationType `json:"operationType"`
	Amount        int64         `json:"amount"`
}

type WalletOperationResponse struct {
	WalletID uuid.UUID `json:"walletId"`
	Balance  int64     `json:"balance"`
}
