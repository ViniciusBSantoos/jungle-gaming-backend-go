package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifySignedAccessToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(
		rand.Reader, 2048,
	)
	if err != nil {
		t.Fatal(err)
	}

	claims := JWTClaims{
		TokenType: "Bearer",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "jogador-123",
			Issuer:  "http://localhost:8081/realms/jungle-gaming",
			Audience: jwt.ClaimStrings{
				"jungle-backend-api",
			},
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(5 * time.Minute),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256, claims,
	)

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	issuer := "http://localhost:8081/realms/jungle-gaming"
	audience := "jungle-backend-api"

	t.Run("AssinaturaValida", func(t *testing.T) {
		result, err := VerifySignedAccessToken(
			signedToken,
			&privateKey.PublicKey,
			issuer,
			audience,
		)

		if err != nil {
			t.Fatalf("Token valido rejeitado: %v", err)
		}

		if result.Subject != "jogador-123" {
			t.Error("Subject incorreto")
		}
	})

	t.Run("AssinaturaInvalida", func(t *testing.T) {
		otherKey, err := rsa.GenerateKey(
			rand.Reader, 2048,
		)
		if err != nil {
			t.Fatal(err)
		}

		_, err = VerifySignedAccessToken(
			signedToken,
			&otherKey.PublicKey,
			issuer,
			audience,
		)

		if err == nil {
			t.Fatal("Assinatura invalida foi aceita")
		}
	})

	t.Run("AudienceIncorreta", func(t *testing.T) {
		_, err := VerifySignedAccessToken(
			signedToken,
			&privateKey.PublicKey,
			issuer,
			"outra-api",
		)

		if err == nil {
			t.Fatal("Audience incorreta foi aceita")
		}
	})
}

func TestVerifySignedAccessTokenExpired(t *testing.T) {
	privateKey, err := rsa.GenerateKey(
		rand.Reader, 2048,
	)
	if err != nil {
		t.Fatal(err)
	}

	issuer := "http://localhost:8081/realms/jungle-gaming"
	audience := "jungle-backend-api"

	claims := JWTClaims{
		TokenType: "Bearer",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "jogador-teste",
			Issuer:  issuer,
			Audience: jwt.ClaimStrings{
				audience,
			},
			IssuedAt: jwt.NewNumericDate(
				time.Now().Add(-10 * time.Minute),
			),
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(-5 * time.Minute),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		claims,
	)

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = VerifySignedAccessToken(
		signedToken,
		&privateKey.PublicKey,
		issuer,
		audience,
	)

	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf(
			"Esperado erro de token expirado, recebido: %v",
			err,
		)
	}
}

func TestVerifySignedAccessTokenWrongIssuer(t *testing.T) {
	privateKey, err := rsa.GenerateKey(
		rand.Reader, 2048,
	)
	if err != nil {
		t.Fatal(err)
	}

	issuer := "http://localhost:8081/realms/jungle-gaming"
	audience := "jungle-backend-api"

	claims := JWTClaims{
		TokenType: "Bearer",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "jogador-teste",
			Issuer:  "http://servidor-falso/realms/outro",
			Audience: jwt.ClaimStrings{
				audience,
			},
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(5 * time.Minute),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		claims,
	)

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = VerifySignedAccessToken(
		signedToken,
		&privateKey.PublicKey,
		issuer,
		audience,
	)

	if !errors.Is(err, jwt.ErrTokenInvalidIssuer) {
		t.Fatalf(
			"Esperado erro de issuer invalido, recebido: %v",
			err,
		)
	}
}
