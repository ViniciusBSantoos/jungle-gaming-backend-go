package service

import (
	"context"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"jungle-gaming-backend-go/internal/database"

	"github.com/google/uuid"
)

func TestBetConcurrencyIntegration(t *testing.T) {
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
		t.Fatal("Use somente o banco de testes local")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 20*time.Second,
	)
	defer cancel()

	pool, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	walletID := uuid.NewString()
	playerID := uuid.NewString()

	// Carteira exclusiva para este teste.
	_, err = pool.Exec(ctx, `
        INSERT INTO wallets (
            id, player_id, currency, balance_cents
        ) VALUES ($1, $2, 'BRL', 10000)
    `, walletID, playerID)
	if err != nil {
		t.Fatal(err)
	}

	betService := NewBetService(pool)

	var wg sync.WaitGroup
	results := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			req := BetRequest{
				WalletID:              walletID,
				PlayerID:              playerID,
				ProviderID:            "provider-001",
				GameID:                "game-001",
				RoundID:               "round-001",
				ExternalTransactionID: uuid.NewString(),
				IdempotencyKey:        uuid.NewString(),
				AmountCents:           8000,
				Currency:              "BRL",
			}

			_, err := betService.ProcessBet(ctx, req)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	success := 0
	rejected := 0

	for err := range results {
		if err == nil {
			success++
		} else if err.Error() == "saldo insuficiente" {
			rejected++
		} else {
			t.Errorf("Erro inesperado: %v", err)
		}
	}

	if success != 1 || rejected != 1 {
		t.Errorf(
			"Esperado 1 aprovada e 1 rejeitada; recebido %d e %d",
			success, rejected,
		)
	}

	var balance int64
	err = pool.QueryRow(ctx,
		"SELECT balance_cents FROM wallets WHERE id = $1",
		walletID,
	).Scan(&balance)
	if err != nil {
		t.Fatal(err)
	}

	if balance != 2000 {
		t.Errorf("Saldo esperado 2000, recebido %d", balance)
	}

	var ledgerCount int
	err = pool.QueryRow(ctx, `
        SELECT COUNT(*) FROM wallet_ledger
        WHERE wallet_id = $1
    `, walletID).Scan(&ledgerCount)
	if err != nil {
		t.Fatal(err)
	}

	if ledgerCount != 1 {
		t.Errorf("Esperado 1 registro no ledger, recebido %d", ledgerCount)
	}

	var outboxCount int
	err = pool.QueryRow(ctx, `
        SELECT COUNT(*) FROM outbox_events
        WHERE aggregate_id = $1
    `, walletID).Scan(&outboxCount)
	if err != nil {
		t.Fatal(err)
	}

	if outboxCount != 1 {
		t.Errorf("Esperado 1 evento na outbox, recebido %d", outboxCount)
	}
}
