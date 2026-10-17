package service

import (
	"context"
	"fmt"

	"jungle-gaming-backend-go/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BetService struct {
	db *pgxpool.Pool
}

func NewBetService(db *pgxpool.Pool) *BetService {
	return &BetService{db: db}
}

func (s *BetService) CheckWalletForBet(
	ctx context.Context,
	walletID string,
	amountCents int64,
) error {
	if amountCents <= 0 {
		return fmt.Errorf("valor da aposta deve ser positivo")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transacao: %w", err)
	}
	defer tx.Rollback(ctx)

	balance, err := repository.LockWallet(ctx, tx, walletID)
	if err != nil {
		return fmt.Errorf("erro ao bloquear carteira: %w", err)
	}

	if balance < amountCents {
		return fmt.Errorf("saldo insuficiente")
	}

	return nil
}

func (s *BetService) ValidateBetWallet(
	ctx context.Context,
	req BetRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transacao: %w", err)
	}
	defer tx.Rollback(ctx)

	wallet, err := repository.LockWalletDetails(
		ctx,
		tx,
		req.WalletID,
	)
	if err != nil {
		return fmt.Errorf("erro ao consultar carteira: %w", err)
	}

	if wallet.PlayerID != req.PlayerID {
		return fmt.Errorf("carteira nao pertence ao jogador")
	}

	if wallet.Currency != req.Currency {
		return fmt.Errorf("moeda da aposta diferente da carteira")
	}

	if wallet.BalanceCents < req.AmountCents {
		return fmt.Errorf("saldo insuficiente")
	}

	return nil
}

func (s *BetService) BeginBetTransaction(
	ctx context.Context,
) (pgx.Tx, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"erro ao iniciar transacao financeira: %w",
			err,
		)
	}

	return tx, nil
}
