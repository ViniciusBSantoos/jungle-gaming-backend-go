package service

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"jungle-gaming-backend-go/internal/database"
)

func TestValidateBetWalletIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL nao configurada")
	}

	if !strings.Contains(databaseURL, "localhost") &&
		!strings.Contains(databaseURL, "127.0.0.1") {
		t.Skip("Teste permitido apenas no banco local")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 10*time.Second,
	)
	defer cancel()

	pool, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	betService := NewBetService(pool)

	req := BetRequest{
		WalletID:              "11111111-1111-4111-8111-111111111111",
		PlayerID:              "22222222-2222-4222-8222-222222222222",
		ProviderID:            "provider-001",
		GameID:                "game-001",
		RoundID:               "round-001",
		ExternalTransactionID: "transaction-001",
		IdempotencyKey:        "key-001",
		AmountCents:           2500,
		Currency:              "BRL",
	}

	if err := betService.ValidateBetWallet(ctx, req); err != nil {
		t.Fatalf("Aposta valida rejeitada: %v", err)
	}

	req.PlayerID = "33333333-3333-4333-8333-333333333333"

	if err := betService.ValidateBetWallet(ctx, req); err == nil {
		t.Error("Deveria rejeitar jogador incorreto")
	}

	req.PlayerID = "22222222-2222-4222-8222-222222222222"
	req.AmountCents = 15000

	if err := betService.ValidateBetWallet(ctx, req); err == nil {
		t.Error("Deveria rejeitar saldo insuficiente")
	}
}
