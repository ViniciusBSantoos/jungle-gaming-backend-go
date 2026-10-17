package domain

import "testing"

func TestNewWallet(t *testing.T) {
	saldo, err := NewMoney(10000, "BRL")
	if err != nil {
		t.Fatal(err)
	}

	carteira, err := NewWallet(
		"wallet-001",
		"player-001",
		saldo,
	)

	if err != nil {
		t.Fatalf("Erro ao criar carteira: %v", err)
	}

	if carteira.Balance.cents != 10000 {
		t.Errorf("Saldo incorreto")
	}

	if carteira.PlayerID != "player-001" {
		t.Errorf("Jogador incorreto")
	}

	if carteira.Version != 1 {
		t.Errorf("Versao inicial incorreta")
	}
}

func TestWalletSemJogador(t *testing.T) {
	saldo, _ := NewMoney(1000, "BRL")

	_, err := NewWallet("wallet-001", "", saldo)

	if err == nil {
		t.Error("Deveria rejeitar carteira sem jogador")
	}
}

func TestWalletSemID(t *testing.T) {
	saldo, _ := NewMoney(1000, "BRL")

	_, err := NewWallet("", "player-001", saldo)

	if err == nil {
		t.Error("Deveria rejeitar carteira sem ID")
	}
}

func TestWalletCredit(t *testing.T) {
	saldo, _ := NewMoney(10000, "BRL")
	carteira, _ := NewWallet("wallet-001", "player-001", saldo)

	credito, _ := NewMoney(5000, "BRL")

	err := carteira.Credit(credito)
	if err != nil {
		t.Fatalf("Erro ao adicionar credito: %v", err)
	}

	if carteira.Balance.cents != 15000 {
		t.Errorf("Esperado 15000, recebido %d", carteira.Balance.cents)
	}

	if carteira.Version != 2 {
		t.Error("Versao deveria ser 2")
	}
}

func TestWalletDebit(t *testing.T) {
	saldo, _ := NewMoney(10000, "BRL")
	carteira, _ := NewWallet("wallet-001", "player-001", saldo)

	aposta, _ := NewMoney(2500, "BRL")

	err := carteira.Debit(aposta)
	if err != nil {
		t.Fatalf("Erro ao debitar: %v", err)
	}

	if carteira.Balance.cents != 7500 {
		t.Errorf("Esperado 7500, recebido %d", carteira.Balance.cents)
	}
}

func TestWalletDebitSaldoInsuficiente(t *testing.T) {
	saldo, _ := NewMoney(1000, "BRL")
	carteira, _ := NewWallet("wallet-001", "player-001", saldo)

	aposta, _ := NewMoney(2500, "BRL")

	err := carteira.Debit(aposta)

	if err == nil {
		t.Error("Deveria retornar erro de saldo insuficiente")
	}

	if carteira.Balance.cents != 1000 {
		t.Error("Saldo nao deveria ser alterado")
	}

	if carteira.Version != 1 {
		t.Error("Versao nao deveria ser alterada")
	}
}
