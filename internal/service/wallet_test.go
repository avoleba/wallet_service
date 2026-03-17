package service

import (
	"context"
	"errors"
	"testing"
	"wallet-api/internal/model"
	"wallet-api/internal/repository"

	"github.com/google/uuid"
)

type mockWalletRepo struct {
	getByID         func(ctx context.Context, id uuid.UUID) (*model.Wallet, error)
	depositOrCreate func(ctx context.Context, id uuid.UUID, amount int64) (*model.Wallet, error)
	withdraw        func(ctx context.Context, id uuid.UUID, amount int64) (*model.Wallet, error)
}

func (m *mockWalletRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, nil
}

func (m *mockWalletRepo) DepositOrCreate(ctx context.Context, id uuid.UUID, amount int64) (*model.Wallet, error) {
	if m.depositOrCreate != nil {
		return m.depositOrCreate(ctx, id, amount)
	}
	return nil, nil
}

func (m *mockWalletRepo) Withdraw(ctx context.Context, id uuid.UUID, amount int64) (*model.Wallet, error) {
	if m.withdraw != nil {
		return m.withdraw(ctx, id, amount)
	}
	return nil, nil
}

func TestWalletService_GetBalance(t *testing.T) {
	ctx := context.Background()
	walletID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	t.Run("returns wallet when found", func(t *testing.T) {
		want := &model.Wallet{ID: walletID, Balance: 100}
		repo := &mockWalletRepo{
			getByID: func(context.Context, uuid.UUID) (*model.Wallet, error) {
				return want, nil
			},
		}
		svc := NewWalletService(repo)
		got, err := svc.GetBalance(ctx, walletID)
		if err != nil {
			t.Fatalf("GetBalance: %v", err)
		}
		if got != want || got.Balance != 100 {
			t.Errorf("got wallet %+v, want %+v", got, want)
		}
	})

	t.Run("returns ErrWalletNotFound when wallet does not exist", func(t *testing.T) {
		repo := &mockWalletRepo{
			getByID: func(context.Context, uuid.UUID) (*model.Wallet, error) {
				return nil, nil
			},
		}
		svc := NewWalletService(repo)
		_, err := svc.GetBalance(ctx, walletID)
		if !errors.Is(err, ErrWalletNotFound) {
			t.Errorf("got err %v, want ErrWalletNotFound", err)
		}
	})

	t.Run("returns error when repo fails", func(t *testing.T) {
		repoErr := errors.New("db error")
		repo := &mockWalletRepo{
			getByID: func(context.Context, uuid.UUID) (*model.Wallet, error) {
				return nil, repoErr
			},
		}
		svc := NewWalletService(repo)
		_, err := svc.GetBalance(ctx, walletID)
		if err != repoErr {
			t.Errorf("got err %v, want %v", err, repoErr)
		}
	})
}

func TestWalletService_ExecuteOperation(t *testing.T) {
	ctx := context.Background()
	walletID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	t.Run("DEPOSIT success", func(t *testing.T) {
		repo := &mockWalletRepo{
			depositOrCreate: func(context.Context, uuid.UUID, int64) (*model.Wallet, error) {
				return &model.Wallet{ID: walletID, Balance: 150}, nil
			},
		}
		svc := NewWalletService(repo)
		resp, err := svc.ExecuteOperation(ctx, model.WalletOperationRequest{
			WalletID:      walletID,
			OperationType: model.OperationDeposit,
			Amount:        100,
		})
		if err != nil {
			t.Fatalf("ExecuteOperation: %v", err)
		}
		if resp.WalletID != walletID || resp.Balance != 150 {
			t.Errorf("got %+v, want WalletID=%s Balance=150", resp, walletID)
		}
	})

	t.Run("WITHDRAW success", func(t *testing.T) {
		repo := &mockWalletRepo{
			withdraw: func(context.Context, uuid.UUID, int64) (*model.Wallet, error) {
				return &model.Wallet{ID: walletID, Balance: 50}, nil
			},
		}
		svc := NewWalletService(repo)
		resp, err := svc.ExecuteOperation(ctx, model.WalletOperationRequest{
			WalletID:      walletID,
			OperationType: model.OperationWithdraw,
			Amount:        50,
		})
		if err != nil {
			t.Fatalf("ExecuteOperation: %v", err)
		}
		if resp.Balance != 50 {
			t.Errorf("got balance %d, want 50", resp.Balance)
		}
	})

	t.Run("amount <= 0 returns ErrInvalidAmount", func(t *testing.T) {
		svc := NewWalletService(&mockWalletRepo{})
		_, err := svc.ExecuteOperation(ctx, model.WalletOperationRequest{
			WalletID:      walletID,
			OperationType: model.OperationDeposit,
			Amount:        0,
		})
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("got err %v, want ErrInvalidAmount", err)
		}
	})

	t.Run("invalid operation type returns ErrInvalidOperation", func(t *testing.T) {
		svc := NewWalletService(&mockWalletRepo{})
		_, err := svc.ExecuteOperation(ctx, model.WalletOperationRequest{
			WalletID:      walletID,
			OperationType: "INVALID",
			Amount:        10,
		})
		if !errors.Is(err, ErrInvalidOperation) {
			t.Errorf("got err %v, want ErrInvalidOperation", err)
		}
	})

	t.Run("WITHDRAW insufficient funds returns ErrInsufficientFunds", func(t *testing.T) {
		repo := &mockWalletRepo{
			withdraw: func(context.Context, uuid.UUID, int64) (*model.Wallet, error) {
				return nil, repository.ErrInsufficientFunds
			},
		}
		svc := NewWalletService(repo)
		_, err := svc.ExecuteOperation(ctx, model.WalletOperationRequest{
			WalletID:      walletID,
			OperationType: model.OperationWithdraw,
			Amount:        100,
		})
		if !errors.Is(err, ErrInsufficientFunds) {
			t.Errorf("got err %v, want ErrInsufficientFunds", err)
		}
	})

	t.Run("WITHDRAW wallet not found returns ErrWalletNotFound", func(t *testing.T) {
		repo := &mockWalletRepo{
			withdraw: func(context.Context, uuid.UUID, int64) (*model.Wallet, error) {
				return nil, nil
			},
		}
		svc := NewWalletService(repo)
		_, err := svc.ExecuteOperation(ctx, model.WalletOperationRequest{
			WalletID:      walletID,
			OperationType: model.OperationWithdraw,
			Amount:        10,
		})
		if !errors.Is(err, ErrWalletNotFound) {
			t.Errorf("got err %v, want ErrWalletNotFound", err)
		}
	})
}
