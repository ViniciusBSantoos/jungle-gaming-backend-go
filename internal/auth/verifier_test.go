package auth

import (
	"context"
	"strings"
	"testing"
)

func TestVerifyAccessTokenEmpty(t *testing.T) {
	verifier := &Verifier{}

	_, err := verifier.VerifyAccessToken(
		context.Background(),
		"",
	)

	if err == nil {
		t.Fatal("Deveria rejeitar token vazio")
	}

	if !strings.Contains(err.Error(), "obrigatorio") {
		t.Errorf("Erro inesperado: %v", err)
	}
}

func TestVerifyAccessTokenSpaces(t *testing.T) {
	verifier := &Verifier{}

	_, err := verifier.VerifyAccessToken(
		context.Background(),
		"   ",
	)

	if err == nil {
		t.Fatal("Deveria rejeitar token com espacos")
	}
}
