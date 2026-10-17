package auth

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRealKeycloakAccessToken(t *testing.T) {
	rawToken := os.Getenv("TEST_ACCESS_TOKEN")

	if rawToken == "" {
		t.Skip("TEST_ACCESS_TOKEN nao configurado")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	issuer := "http://localhost:8081/realms/jungle-gaming"

	claims, err := VerifyAccessTokenWithJWKS(
		ctx,
		rawToken,
		issuer+"/protocol/openid-connect/certs",
		issuer,
		"jungle-backend-api",
	)
	if err != nil {
		t.Fatalf("Token real rejeitado: %v", err)
	}

	if claims.Subject == "" {
		t.Fatal("Token sem identificacao do usuario")
	}

	if !claims.HasRole("jungle-backend-api", "wallet:read") {
		t.Fatal("Permissao wallet:read nao encontrada")
	}

	if !claims.HasRole("jungle-backend-api", "wallet:write") {
		t.Fatal("Permissao wallet:write nao encontrada")
	}

	t.Log("Access token real validado com sucesso!")
}
