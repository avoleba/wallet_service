package repository

import (
	"context"
	"database/sql"
	"errors"
	"wallet-api/internal/model"

	"github.com/google/uuid"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	var w model.Wallet
	err := r.db.QueryRowContext(ctx,
		`SELECT id, balance, created_at, updated_at FROM wallets WHERE id = $1`,
		id.String(),
	).Scan(&w.ID, &w.Balance, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WalletRepository) Create(ctx context.Context, id uuid.UUID, initialBalance int64) (*model.Wallet, error) {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO wallets (id, balance, created_at, updated_at) VALUES ($1::uuid, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		id.String(), initialBalance,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, id uuid.UUID, newBalance int64) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE wallets SET balance = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2::uuid`,
		newBalance, id.String(),
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *WalletRepository) DepositOrCreate(ctx context.Context, id uuid.UUID, amount int64) (*model.Wallet, error) {
	var w model.Wallet
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO wallets (id, balance, created_at, updated_at)
		VALUES ($1::uuid, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			balance = wallets.balance + $2,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, balance, created_at, updated_at`,
		id.String(), amount,
	).Scan(&w.ID, &w.Balance, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WalletRepository) Withdraw(ctx context.Context, id uuid.UUID, amount int64) (*model.Wallet, error) {
	var w model.Wallet
	err := r.db.QueryRowContext(ctx, `
		UPDATE wallets SET balance = balance - $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2::uuid AND balance >= $1
		RETURNING id, balance, created_at, updated_at`,
		amount, id.String(),
	).Scan(&w.ID, &w.Balance, &w.CreatedAt, &w.UpdatedAt)
	if err == nil {
		return &w, nil
	}
	if err == sql.ErrNoRows {
		exists, errExists := r.exists(ctx, id)
		if errExists != nil {
			return nil, errExists
		}
		if !exists {
			return nil, nil
		}
			return nil, ErrInsufficientFunds
	}
	return nil, err
}

var ErrInsufficientFunds = errors.New("insufficient funds")

func (r *WalletRepository) exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT 1 FROM wallets WHERE id = $1::uuid`, id.String()).Scan(&n)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *WalletRepository) Migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS wallets (
			id UUID PRIMARY KEY,
			balance BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}
