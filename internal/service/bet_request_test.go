package service

import "testing"

func validBetRequest() BetRequest {
	return BetRequest{
		WalletID:              "wallet-001",
		PlayerID:              "player-001",
		ProviderID:            "provider-001",
		GameID:                "game-001",
		RoundID:               "round-001",
		ExternalTransactionID: "transaction-001",
		IdempotencyKey:        "key-001",
		AmountCents:           2500,
		Currency:              "BRL",
	}
}

func TestBetRequestValid(t *testing.T) {
	req := validBetRequest()

	if err := req.Validate(); err != nil {
		t.Fatalf("Aposta valida rejeitada: %v", err)
	}
}

func TestBetRequestMissingPlayer(t *testing.T) {
	req := validBetRequest()
	req.PlayerID = ""

	if err := req.Validate(); err == nil {
		t.Error("Deveria rejeitar jogador vazio")
	}
}

func TestBetRequestInvalidAmount(t *testing.T) {
	req := validBetRequest()
	req.AmountCents = -100

	if err := req.Validate(); err == nil {
		t.Error("Deveria rejeitar valor negativo")
	}
}

func TestBetRequestInvalidCurrency(t *testing.T) {
	req := validBetRequest()
	req.Currency = "USD"

	if err := req.Validate(); err == nil {
		t.Error("Deveria rejeitar moeda nao suportada")
	}
}

func TestBetRequestMissingIdempotencyKey(t *testing.T) {
	req := validBetRequest()
	req.IdempotencyKey = ""

	if err := req.Validate(); err == nil {
		t.Error("Deveria rejeitar chave de idempotencia vazia")
	}
}
