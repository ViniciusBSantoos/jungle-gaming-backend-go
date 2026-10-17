package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestFetchJWKSIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	url := "http://localhost:8081/realms/jungle-gaming/protocol/openid-connect/certs"

	keys, err := FetchJWKS(ctx, url)
	if err != nil {
		t.Fatalf("Erro ao buscar chaves: %v", err)
	}

	if len(keys.Keys) == 0 {
		t.Fatal("Nenhuma chave publica encontrada")
	}

	found := false

	for _, key := range keys.Keys {
		if key.Kty == "RSA" &&
			key.Kid != "" &&
			key.N != "" &&
			key.E != "" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Nenhuma chave publica RSA valida encontrada")
	}

	t.Log("Chaves publicas do Keycloak encontradas!")
}

func TestPublicKeyByKIDIntegration(t *testing.T) {
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

	// Procura uma chave RSA para assinatura RS256.
	var kid string

	for _, key := range keys.Keys {
		if key.Kty == "RSA" &&
			key.Use == "sig" &&
			key.Alg == "RS256" &&
			key.Kid != "" {
			kid = key.Kid
			break
		}
	}

	if kid == "" {
		t.Fatal("Nenhuma chave RS256 encontrada")
	}

	// Verifica se conseguimos selecionar a chave.
	publicKey, err := keys.PublicKeyByKID(kid)
	if err != nil {
		t.Fatalf("Erro ao selecionar chave: %v", err)
	}

	if publicKey == nil {
		t.Fatal("Chave publica nao encontrada")
	}

	// Um identificador inexistente deve ser rejeitado.
	_, err = keys.PublicKeyByKID("kid-inexistente")

	if err == nil {
		t.Fatal("KID inexistente deveria ser rejeitado")
	}

	t.Log("Selecao de chave RSA por KID funcionando!")
}

func TestJWKSRejectsUnknownSignature(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	claims := JWTClaims{
		TokenType: "Bearer",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  "jogador-teste",
			Issuer:   "http://localhost:8081/realms/jungle-gaming",
			Audience: jwt.ClaimStrings{"jungle-backend-api"},
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(5 * time.Minute),
			),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "chave-desconhecida"

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = VerifyAccessTokenWithJWKS(
		ctx,
		signedToken,
		"http://localhost:8081/realms/jungle-gaming/protocol/openid-connect/certs",
		"http://localhost:8081/realms/jungle-gaming",
		"jungle-backend-api",
	)

	if err == nil {
		t.Fatal("Token com chave desconhecida deveria ser rejeitado")
	}
}

func TestJWKSRejectsInvalidSignature(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	issuer := "http://localhost:8081/realms/jungle-gaming"
	jwksURL := issuer + "/protocol/openid-connect/certs"

	keys, err := FetchJWKS(ctx, jwksURL)
	if err != nil {
		t.Fatal(err)
	}

	// Encontra um KID verdadeiro do Keycloak.
	var validKID string

	for _, key := range keys.Keys {
		if key.Kty == "RSA" &&
			key.Use == "sig" &&
			key.Alg == "RS256" &&
			key.Kid != "" {
			validKID = key.Kid
			break
		}
	}

	if validKID == "" {
		t.Fatal("Nenhuma chave RS256 encontrada")
	}

	// Gera uma chave falsa.
	fakeKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	claims := JWTClaims{
		TokenType: "Bearer",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "jogador-teste",
			Issuer:  issuer,
			Audience: jwt.ClaimStrings{
				"jungle-backend-api",
			},
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(5 * time.Minute),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		claims,
	)

	// Usa um KID verdadeiro com assinatura falsa.
	token.Header["kid"] = validKID

	signedToken, err := token.SignedString(fakeKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = VerifyAccessTokenWithJWKS(
		ctx,
		signedToken,
		jwksURL,
		issuer,
		"jungle-backend-api",
	)

	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Fatalf(
			"Esperado erro de assinatura invalida, recebido: %v",
			err,
		)
	}
}
