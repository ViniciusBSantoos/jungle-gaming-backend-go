package service

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"jungle-gaming-backend-go/internal/database"

	"github.com/google/uuid"
)

func TestBetRollbackIntegration(t *testing.T) {
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
		t.Fatal("Use apenas o banco local de testes")
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

	_, err = pool.Exec(ctx, `
    CREATE OR REPLACE FUNCTION fail_outbox_test()
    RETURNS TRIGGER AS $$
    BEGIN
        RAISE EXCEPTION 'falha simulada na outbox';
    END;
    $$ LANGUAGE plpgsql;

    CREATE TRIGGER fail_outbox_insert
    BEFORE INSERT ON outbox_events
    FOR EACH ROW
    EXECUTE FUNCTION fail_outbox_test();
`)
	if err != nil {
		t.Fatalf("Erro ao preparar falha simulada: %v", err)
	}

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(), 5*time.Second,
		)
		defer cancel()

		_, cleanupErr := pool.Exec(cleanupCtx, `
        DROP TRIGGER IF EXISTS fail_outbox_insert
        ON outbox_events;

        DROP FUNCTION IF EXISTS fail_outbox_test();
    `)
		if cleanupErr != nil {
			t.Errorf("Erro ao remover trigger: %v", cleanupErr)
		}
	}()

	walletID := uuid.NewString()
	playerID := uuid.NewString()
	externalID := uuid.NewString()
	idempotencyKey := uuid.NewString()

	_, err = pool.Exec(ctx, `
        INSERT INTO wallets
        (id, player_id, currency, balance_cents)
        VALUES ($1, $2, 'BRL', 10000)
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
		ExternalTransactionID: externalID,
		IdempotencyKey:        idempotencyKey,
		AmountCents:           2500,
		Currency:              "BRL",
	}

	betService := NewBetService(pool)

	_, err = betService.ProcessBet(ctx, req)

	if err == nil {
		t.Fatal("Era esperado erro na outbox")
	}

	if !strings.Contains(err.Error(), "falha simulada na outbox") {
		t.Fatalf("Erro inesperado: %v", err)
	}

	var balance int64
	err = pool.QueryRow(ctx,
		"SELECT balance_cents FROM wallets WHERE id = $1",
		walletID,
	).Scan(&balance)
	if err != nil {
		t.Fatal(err)
	}

	if balance != 10000 {
		t.Errorf("Saldo alterado indevidamente: %d", balance)
	}

	var transactions int
	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM wager_transactions WHERE wallet_id = $1",
		walletID,
	).Scan(&transactions)
	if err != nil {
		t.Fatal(err)
	}

	if transactions != 0 {
		t.Errorf("Transacao parcial encontrada: %d", transactions)
	}

	var ledger int
	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM wallet_ledger WHERE wallet_id = $1",
		walletID,
	).Scan(&ledger)
	if err != nil {
		t.Fatal(err)
	}

	if ledger != 0 {
		t.Errorf("Ledger parcial encontrado: %d", ledger)
	}

	var outbox int
	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = $1",
		walletID,
	).Scan(&outbox)
	if err != nil {
		t.Fatal(err)
	}

	if outbox != 0 {
		t.Errorf("Outbox parcial encontrada: %d", outbox)
	}
}
