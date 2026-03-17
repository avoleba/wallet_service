package repository

import (
	"context"
	"testing"
	"time"
	"wallet-api/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var testTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestWalletRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repo := NewWalletRepository(db)
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	t.Run("returns wallet when found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "balance", "created_at", "updated_at"}).
			AddRow(id.String(), int64(100), testTime, testTime)
		mock.ExpectQuery(`SELECT id, balance, created_at, updated_at FROM wallets WHERE id = \$1`).
			WithArgs(id.String()).
			WillReturnRows(rows)

		w, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if w == nil || w.ID != id || w.Balance != 100 {
			t.Errorf("got %+v, want wallet with id=%s balance=100", w, id)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})

	t.Run("returns nil when not found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "balance", "created_at", "updated_at"})
		mock.ExpectQuery(`SELECT id, balance, created_at, updated_at FROM wallets WHERE id = \$1`).
			WithArgs(id.String()).
			WillReturnRows(rows)

		w, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if w != nil {
			t.Errorf("got wallet %+v, want nil", w)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})
}

func TestWalletRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repo := NewWalletRepository(db)
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	mock.ExpectExec(`INSERT INTO wallets`).
		WithArgs(id.String(), int64(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	rows := sqlmock.NewRows([]string{"id", "balance", "created_at", "updated_at"}).
		AddRow(id.String(), int64(0), testTime, testTime)
	mock.ExpectQuery(`SELECT id, balance, created_at, updated_at FROM wallets WHERE id = \$1`).
		WithArgs(id.String()).
		WillReturnRows(rows)

	w, err := repo.Create(ctx, id, 0)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if w == nil || w.Balance != 0 {
		t.Errorf("got %+v", w)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestWalletRepository_UpdateBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repo := NewWalletRepository(db)
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	mock.ExpectExec(`UPDATE wallets SET balance`).
		WithArgs(int64(50), id.String()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateBalance(ctx, id, 50)
	if err != nil {
		t.Fatalf("UpdateBalance: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestWalletRepository_DepositOrCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repo := NewWalletRepository(db)
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	amount := int64(100)

	rows := sqlmock.NewRows([]string{"id", "balance", "created_at", "updated_at"}).
		AddRow(id.String(), amount, testTime, testTime)
	mock.ExpectQuery(`INSERT INTO wallets`).
		WithArgs(id.String(), amount).
		WillReturnRows(rows)

	w, err := repo.DepositOrCreate(ctx, id, amount)
	if err != nil {
		t.Fatalf("DepositOrCreate: %v", err)
	}
	if w == nil || w.Balance != amount {
		t.Errorf("got %+v, want balance=%d", w, amount)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestWalletRepository_Withdraw(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repo := NewWalletRepository(db)
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	amount := int64(30)

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "balance", "created_at", "updated_at"}).
			AddRow(id.String(), int64(70), testTime, testTime)
		mock.ExpectQuery(`UPDATE wallets SET balance = balance`).
			WithArgs(amount, id.String()).
			WillReturnRows(rows)

		w, err := repo.Withdraw(ctx, id, amount)
		if err != nil {
			t.Fatalf("Withdraw: %v", err)
		}
		if w == nil || w.Balance != 70 {
			t.Errorf("got %+v, want balance=70", w)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})
}

func TestWalletRepository_Migrate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repo := NewWalletRepository(db)

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS wallets`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Migrate(ctx)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// Ensure *WalletRepository implements WalletRepository interface used by service
var _ interface {
	GetByID(context.Context, uuid.UUID) (*model.Wallet, error)
	DepositOrCreate(context.Context, uuid.UUID, int64) (*model.Wallet, error)
	Withdraw(context.Context, uuid.UUID, int64) (*model.Wallet, error)
} = (*WalletRepository)(nil)
