package service

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"jungle-gaming-backend-go/internal/database"

	"github.com/google/uuid"
)

func TestBetIdempotencyIntegration(t *testing.T) {
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
		t.Fatal("Use somente o banco local de testes")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 15*time.Second,
	)
	defer cancel()

	pool, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}

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

	betService := NewBetService(pool)

	first, err := betService.ProcessBet(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	second, err := betService.ProcessBet(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if !second.Replayed {
		t.Error("A segunda aposta deveria ser replay")
	}

	if first.TransactionID != second.TransactionID {
		t.Error("A aposta repetida deveria manter o mesmo ID")
	}

	if first.BalanceCents != second.BalanceCents {
		t.Error("Saldo do replay deveria ser igual")
	}

	var balance int64
	err = pool.QueryRow(ctx,
		"SELECT balance_cents FROM wallets WHERE id = $1",
		walletID,
	).Scan(&balance)
	if err != nil {
		t.Fatal(err)
	}

	if balance != 7500 {
		t.Errorf("Saldo esperado 7500, recebido %d", balance)
	}
}

func TestBetIdempotencyConflict(t *testing.T) {
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
		t.Fatal("Use somente o banco local de testes")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 15*time.Second,
	)
	defer cancel()

	pool, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}

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

	betService := NewBetService(pool)

	_, err = betService.ProcessBet(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	// Tenta reutilizar a mesma chave com outro valor.
	req.AmountCents = 3000

	_, err = betService.ProcessBet(ctx, req)
	if err == nil {
		t.Fatal("Deveria rejeitar dados diferentes")
	}

	var balance int64
	err = pool.QueryRow(ctx,
		"SELECT balance_cents FROM wallets WHERE id = $1",
		walletID,
	).Scan(&balance)
	if err != nil {
		t.Fatal(err)
	}

	if balance != 7500 {
		t.Errorf("Saldo esperado 7500, recebido %d", balance)
	}
}
