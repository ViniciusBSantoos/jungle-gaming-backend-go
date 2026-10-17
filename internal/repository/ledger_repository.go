package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func InsertLedgerEntry(
	ctx context.Context,
	tx pgx.Tx,
	id string,
	walletID string,
	transactionID string,
	amountCents int64,
	balanceBefore int64,
	balanceAfter int64,
) error {
	if amountCents <= 0 {
		return fmt.Errorf("valor deve ser positivo")
	}

	if balanceBefore < amountCents ||
		balanceAfter != balanceBefore-amountCents {
		return fmt.Errorf("saldos inconsistentes")
	}

	_, err := tx.Exec(
		ctx,
		`INSERT INTO wallet_ledger (
            id,
            wallet_id,
            transaction_id,
            direction,
            amount_cents,
            balance_before_cents,
            balance_after_cents
        ) VALUES (
            $1, $2, $3, 'DEBIT',
            $4, $5, $6
        )`,
		id,
		walletID,
		transactionID,
		amountCents,
		balanceBefore,
		balanceAfter,
	)

	return err
}
