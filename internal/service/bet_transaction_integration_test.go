package service

import (
	"context"
	"jungle-gaming-backend-go/internal/database"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProcessBetIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")

	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL nao configurada")
	}

	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.Path != "/jungle_gaming_test" ||
		(parsed.Hostname() != "localhost" &&
			parsed.Hostname() != "127.0.0.1") {
		t.Fatal("Use somente o banco local jungle_gaming_test")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 15*time.Second,
	)
	defer cancel()

	pool, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)

		walletID := uuid.NewString()
		playerID := uuid.NewString()

		_, err = pool.Exec(ctx, `
    INSERT INTO wallets (
        id, player_id, currency, balance_cents
    ) VALUES ($1, $2, 'BRL', 10000)
`, walletID, playerID)

		if err != nil {
			t.Fatalf("Erro ao criar carteira de teste: %v", err)
		}

	}
	defer pool.Close()

	walletID := uuid.NewString()
	playerID := uuid.NewString()

	_, err = pool.Exec(ctx, `
    INSERT INTO wallets (
        id, player_id, currency, balance_cents
    ) VALUES ($1, $2, 'BRL', 10000)
`, walletID, playerID)

	if err != nil {
		t.Fatalf("Erro ao criar carteira de teste: %v", err)
	}

	service := NewBetService(pool)

	req := BetRequest{
		WalletID:              walletID,
		PlayerID:              playerID,
		ProviderID:            "provider-001",
		GameID:                "game-001",
		RoundID:               "round-001",
		ExternalTransactionID: uuid.NewString(),
		IdempotencyKey:        uuid.NewString(),
		AmountCents:           2500,
		Currency:              "BRL",
	}

	result, err := service.ProcessBet(ctx, req)
	if err != nil {
		t.Fatalf("Erro ao processar aposta: %v", err)
	}

	if result.Status != "PROCESSED" {
		t.Errorf("Status incorreto: %s", result.Status)
	}

	var balance int64
	err = pool.QueryRow(
		ctx,
		"SELECT balance_cents FROM wallets WHERE id = $1",
		req.WalletID,
	).Scan(&balance)

	if err != nil {
		t.Fatal(err)
	}

	if balance != 7500 {
		t.Errorf(
			"Saldo esperado 7500, recebido %d",
			balance,
		)
	}

	if balance != result.BalanceCents {
		t.Errorf("Saldo do banco diferente do resultado")
	}

	var ledgerCount int
	err = pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM wallet_ledger WHERE transaction_id = $1",
		result.TransactionID,
	).Scan(&ledgerCount)
	if err != nil {
		t.Fatal(err)
	}

	if ledgerCount != 1 {
		t.Error("Aposta deveria gerar um registro no ledger")
	}

	var outboxCount int
	err = pool.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM outbox_events
         WHERE payload->>'transaction_id' = $1`,
		result.TransactionID,
	).Scan(&outboxCount)
	if err != nil {
		t.Fatal(err)
	}

	if outboxCount != 1 {
		t.Error("Aposta deveria gerar um evento na outbox")
	}
}
