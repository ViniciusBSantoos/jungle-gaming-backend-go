package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type StoredBet struct {
	ID            string
	PayloadHash   string
	Status        string
	FailureCode   *string
	ResultBalance *int64
}

func FindBetByIdempotencyKey(
	ctx context.Context,
	tx pgx.Tx,
	key string,
) (StoredBet, error) {
	var bet StoredBet

	err := tx.QueryRow(
		ctx,
		`SELECT id, payload_hash, status,
                failure_code, result_balance_cents
         FROM wager_transactions
         WHERE idempotency_key = $1`,
		key,
	).Scan(
		&bet.ID,
		&bet.PayloadHash,
		&bet.Status,
		&bet.FailureCode,
		&bet.ResultBalance,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return StoredBet{}, pgx.ErrNoRows
	}

	return bet, err
}

func InsertBet(
	ctx context.Context,
	tx pgx.Tx,
	id string,
	walletID string,
	playerID string,
	providerID string,
	gameID string,
	roundID string,
	externalID string,
	idempotencyKey string,
	payloadHash string,
	amountCents int64,
	currency string,
) error {
	_, err := tx.Exec(
		ctx,
		`INSERT INTO wager_transactions (
            id, wallet_id, player_id,
            provider_id, game_id, round_id,
            external_transaction_id, idempotency_key,
            payload_hash, transaction_type,
            amount_cents, currency, status
        ) VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9, 'BET',
            $10, $11, 'PENDING'
        )`,
		id, walletID, playerID,
		providerID, gameID, roundID,
		externalID, idempotencyKey,
		payloadHash, amountCents, currency,
	)

	return err
}

func CompleteBet(
	ctx context.Context,
	tx pgx.Tx,
	transactionID string,
	balanceCents int64,
) error {
	result, err := tx.Exec(
		ctx,
		`UPDATE wager_transactions
         SET status = 'PROCESSED',
             result_balance_cents = $1,
             updated_at = NOW()
         WHERE id = $2
           AND status = 'PENDING'`,
		balanceCents,
		transactionID,
	)

	if err != nil {
		return fmt.Errorf(
			"erro ao concluir aposta: %w", err,
		)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf(
			"transacao inexistente ou nao pendente",
		)
	}

	return nil
}
