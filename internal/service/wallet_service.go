package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletService struct {
	db *pgxpool.Pool
}

func NewWalletService(db *pgxpool.Pool) *WalletService {
	return &WalletService{
		db: db,
	}
}

func (s *WalletService) CheckBalance(
	ctx context.Context,
	walletID string,
) (int64, error) {
	var balance int64

	err := s.db.QueryRow(
		ctx,
		`SELECT balance_cents
         FROM wallets
         WHERE id = $1`,
		walletID,
	).Scan(&balance)

	return balance, err
}
