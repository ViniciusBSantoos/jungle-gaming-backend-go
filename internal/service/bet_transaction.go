package service

import (
	"context"
	"fmt"

	"jungle-gaming-backend-go/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *BetService) ProcessBet(
	ctx context.Context,
	req BetRequest,
) (BetResult, error) {
	if err := req.Validate(); err != nil {
		return BetResult{}, err
	}

	hash, err := req.PayloadHash()
	if err != nil {
		return BetResult{}, err
	}

	tx, err := s.BeginBetTransaction(ctx)
	if err != nil {
		return BetResult{}, err
	}
	defer tx.Rollback(ctx)

	wallet, err := repository.LockWalletDetails(
		ctx, tx, req.WalletID,
	)
	if err != nil {
		return BetResult{}, err
	}

	if wallet.PlayerID != req.PlayerID {
		return BetResult{}, fmt.Errorf("jogador invalido")
	}

	if wallet.Currency != req.Currency {
		return BetResult{}, fmt.Errorf("moeda invalida")
	}

	stored, err := repository.FindBetByIdempotencyKey(
		ctx, tx, req.IdempotencyKey,
	)

	if err == nil {
		if stored.PayloadHash != hash {
			return BetResult{}, fmt.Errorf(
				"chave reutilizada com dados diferentes",
			)
		}

		if stored.Status != "PROCESSED" ||
			stored.ResultBalance == nil {
			return BetResult{}, fmt.Errorf(
				"transacao ainda nao concluida",
			)
		}

		return BetResult{
			TransactionID: stored.ID,
			Status:        stored.Status,
			BalanceCents:  *stored.ResultBalance,
			Replayed:      true,
		}, nil
	}

	if err != pgx.ErrNoRows {
		return BetResult{}, err
	}

	if wallet.BalanceCents < req.AmountCents {
		return BetResult{}, fmt.Errorf("saldo insuficiente")
	}

	newBalance := wallet.BalanceCents - req.AmountCents
	transactionID := uuid.NewString()

	err = repository.InsertBet(
		ctx, tx, transactionID,
		req.WalletID, req.PlayerID,
		req.ProviderID, req.GameID, req.RoundID,
		req.ExternalTransactionID,
		req.IdempotencyKey, hash,
		req.AmountCents, req.Currency,
	)
	if err != nil {
		return BetResult{}, err
	}

	err = repository.UpdateWalletBalance(
		ctx, tx, req.WalletID, newBalance,
	)
	if err != nil {
		return BetResult{}, err
	}

	err = repository.InsertLedgerEntry(
		ctx, tx, uuid.NewString(),
		req.WalletID, transactionID,
		req.AmountCents,
		wallet.BalanceCents, newBalance,
	)
	if err != nil {
		return BetResult{}, err
	}

	payload := map[string]any{
		"transaction_id": transactionID,
		"wallet_id":      req.WalletID,
		"amount_cents":   req.AmountCents,
		"balance_cents":  newBalance,
		"currency":       req.Currency,
	}

	err = repository.InsertOutboxEvent(
		ctx, tx, uuid.NewString(),
		req.WalletID, "BET_PROCESSED", payload,
	)
	if err != nil {
		return BetResult{}, err
	}

	err = repository.CompleteBet(
		ctx, tx, transactionID, newBalance,
	)
	if err != nil {
		return BetResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return BetResult{}, fmt.Errorf(
			"erro ao confirmar transacao: %w", err,
		)
	}

	return BetResult{
		TransactionID: transactionID,
		Status:        "PROCESSED",
		BalanceCents:  newBalance,
		Replayed:      false,
	}, nil
}
