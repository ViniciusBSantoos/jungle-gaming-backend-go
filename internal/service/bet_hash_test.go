package service

import "testing"

func TestPayloadHashIgual(t *testing.T) {
	aposta1 := validBetRequest()
	aposta2 := validBetRequest()

	hash1, err := aposta1.PayloadHash()
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := aposta2.PayloadHash()
	if err != nil {
		t.Fatal(err)
	}

	if hash1 != hash2 {
		t.Error("Apostas iguais devem gerar o mesmo hash")
	}
}

func TestPayloadHashDiferente(t *testing.T) {
	aposta1 := validBetRequest()
	aposta2 := validBetRequest()

	aposta2.AmountCents = 5000

	hash1, err := aposta1.PayloadHash()
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := aposta2.PayloadHash()
	if err != nil {
		t.Fatal(err)
	}

	if hash1 == hash2 {
		t.Error("Apostas diferentes nao devem gerar o mesmo hash")
	}
}
