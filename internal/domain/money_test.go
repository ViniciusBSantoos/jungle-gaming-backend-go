package domain

import "testing"

func TestNewMoney(t *testing.T) {
	dinheiro, err := NewMoney(2500, "BRL")

	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	if dinheiro.cents != 2500 {
		t.Errorf("Esperado 2500, recebido %d", dinheiro.cents)
	}
}

func TestMoneyNegativo(t *testing.T) {
	_, err := NewMoney(-100, "BRL")

	if err == nil {
		t.Error("Deveria rejeitar valor negativo")
	}
}

func TestMoneyAdd(t *testing.T) {
	valor1, _ := NewMoney(10000, "BRL")
	valor2, _ := NewMoney(2500, "BRL")

	resultado, err := valor1.Add(valor2)

	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	if resultado.cents != 12500 {
		t.Errorf("Esperado 12500, recebido %d", resultado.cents)
	}
}

func TestMoneySubtract(t *testing.T) {
	saldo, _ := NewMoney(10000, "BRL")
	aposta, _ := NewMoney(2500, "BRL")

	resultado, err := saldo.Subtract(aposta)

	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	if resultado.cents != 7500 {
		t.Errorf("Esperado 7500, recebido %d", resultado.cents)
	}
}

func TestMoneySaldoInsuficiente(t *testing.T) {
	saldo, _ := NewMoney(1000, "BRL")
	aposta, _ := NewMoney(2500, "BRL")

	_, err := saldo.Subtract(aposta)

	if err == nil {
		t.Error("Deveria retornar erro de saldo insuficiente")
	}
}

func TestMoneyMoedaInvalida(t *testing.T) {
	_, err := NewMoney(1000, "USD")

	if err == nil {
		t.Error("Deveria rejeitar moeda nao suportada")
	}
}

func TestMoneyOverflow(t *testing.T) {
	valor1, _ := NewMoney(1<<63-1, "BRL")
	valor2, _ := NewMoney(1, "BRL")

	_, err := valor1.Add(valor2)

	if err == nil {
		t.Error("Deveria impedir overflow")
	}
}
