package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepository struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{
		db: db,
	}
}

func (r *WalletRepository) CountWallets(
	ctx context.Context,
) (int64, error) {
	var total int64

	err := r.db.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM wallets",
	).Scan(&total)

	return total, err
}

func LockWallet(
	ctx context.Context,
	tx pgx.Tx,
	walletID string,
) (int64, error) {
	var balance int64

	err := tx.QueryRow(
		ctx,
		`SELECT balance_cents
         FROM wallets
         WHERE id = $1
         FOR UPDATE`,
		walletID,
	).Scan(&balance)

	return balance, err
}

type LockedWallet struct {
	PlayerID     string
	Currency     string
	BalanceCents int64
}

func LockWalletDetails(
	ctx context.Context,
	tx pgx.Tx,
	walletID string,
) (LockedWallet, error) {
	var wallet LockedWallet

	err := tx.QueryRow(
		ctx,
		`SELECT player_id, currency, balance_cents
         FROM wallets
         WHERE id = $1
         FOR UPDATE`,
		walletID,
	).Scan(
		&wallet.PlayerID,
		&wallet.Currency,
		&wallet.BalanceCents,
	)

	return wallet, err
}

func UpdateWalletBalance(
	ctx context.Context,
	tx pgx.Tx,
	walletID string,
	newBalance int64,
) error {
	if newBalance < 0 {
		return fmt.Errorf("saldo negativo nao permitido")
	}

	result, err := tx.Exec(
		ctx,
		`UPDATE wallets
         SET balance_cents = $1,
             version = version + 1,
             updated_at = NOW()
         WHERE id = $2`,
		newBalance,
		walletID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("carteira nao encontrada")
	}

	return nil
}

type PlayerWallet struct {
	ID           string
	PlayerID     string
	Currency     string
	BalanceCents int64
}

func (r *WalletRepository) FindByPlayerID(
	ctx context.Context,
	playerID string,
) (PlayerWallet, error) {
	var wallet PlayerWallet

	err := r.db.QueryRow(
		ctx,
		`SELECT id, player_id, currency, balance_cents
         FROM wallets
         WHERE player_id = $1
           AND currency = 'BRL'`,
		playerID,
	).Scan(
		&wallet.ID,
		&wallet.PlayerID,
		&wallet.Currency,
		&wallet.BalanceCents,
	)

	if err != nil {
		return PlayerWallet{}, err
	}

	return wallet, nil
}
