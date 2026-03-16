package service

import (
	"context"
	"errors"
	"wallet-api/internal/model"
	"wallet-api/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrWalletNotFound   = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds for withdrawal")
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrInvalidOperation  = errors.New("operationType must be DEPOSIT or WITHDRAW")
)

type WalletService struct {
	repo *repository.WalletRepository
}

func NewWalletService(repo *repository.WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error) {
	wallet, err := s.repo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, ErrWalletNotFound
	}
	return wallet, nil
}

func (s *WalletService) ExecuteOperation(ctx context.Context, req model.WalletOperationRequest) (*model.WalletOperationResponse, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if req.OperationType != model.OperationDeposit && req.OperationType != model.OperationWithdraw {
		return nil, ErrInvalidOperation
	}

	switch req.OperationType {
	case model.OperationDeposit:
		wallet, err := s.repo.DepositOrCreate(ctx, req.WalletID, req.Amount)
		if err != nil {
			return nil, err
		}
		return &model.WalletOperationResponse{WalletID: wallet.ID, Balance: wallet.Balance}, nil
	case model.OperationWithdraw:
		wallet, err := s.repo.Withdraw(ctx, req.WalletID, req.Amount)
		if err != nil {
			if errors.Is(err, repository.ErrInsufficientFunds) {
				return nil, ErrInsufficientFunds
			}
			return nil, err
		}
		if wallet == nil {
			return nil, ErrWalletNotFound
		}
		return &model.WalletOperationResponse{WalletID: wallet.ID, Balance: wallet.Balance}, nil
	}
	return nil, ErrInvalidOperation
}
