package auth

import (
	"context"
	"testing"
	"time"
)

func TestKeycloakConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	config, err := NewConfig(
		"http://localhost:8081/realms/jungle-gaming",
		"jungle-backend-api",
	)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := NewVerifier(ctx, config)
	if err != nil {
		t.Fatalf("Erro ao conectar ao Keycloak: %v", err)
	}

	if verifier == nil {
		t.Fatal("Verificador nao foi criado")
	}

	t.Log("Conexao OIDC com Keycloak realizada com sucesso!")
}

func TestKeycloakInvalidToken(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	config, err := NewConfig(
		"http://localhost:8081/realms/jungle-gaming",
		"jungle-backend-api",
	)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := NewVerifier(ctx, config)
	if err != nil {
		t.Fatal(err)
	}

	_, err = verifier.VerifyAccessToken(
		ctx,
		"token.invalido.teste",
	)

	if err == nil {
		t.Fatal("Token invalido deveria ser rejeitado")
	}
}

func TestJWKPublicKeyIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	jwksURL := "http://localhost:8081/realms/jungle-gaming/protocol/openid-connect/certs"

	keys, err := FetchJWKS(ctx, jwksURL)
	if err != nil {
		t.Fatalf("Erro ao buscar chaves: %v", err)
	}

	found := false

	for _, key := range keys.Keys {
		if key.Kty != "RSA" ||
			key.Use != "sig" ||
			key.Kid == "" {
			continue
		}

		publicKey, err := key.PublicKey()
		if err != nil {
			t.Fatalf("Erro ao converter chave RSA: %v", err)
		}

		if publicKey.N.BitLen() < 2048 {
			t.Fatal("Chave RSA menor que 2048 bits")
		}

		if publicKey.E < 3 {
			t.Fatal("Expoente RSA invalido")
		}

		found = true
		break
	}

	if !found {
		t.Fatal("Nenhuma chave RSA de assinatura encontrada")
	}

	t.Log("Chave publica RSA convertida com sucesso!")
}
