package auth

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	TokenType      string                 `json:"typ"`
	PreferredUser  string                 `json:"preferred_username"`
	ResourceAccess map[string]ClientRoles `json:"resource_access"`

	jwt.RegisteredClaims
}

func (c JWTClaims) Validate() error {
	if c.TokenType != "Bearer" {
		return fmt.Errorf("tipo de token invalido")
	}

	if c.Subject == "" {
		return fmt.Errorf("subject obrigatorio")
	}

	return nil
}

func VerifySignedAccessToken(
	rawToken string,
	publicKey *rsa.PublicKey,
	issuer string,
	audience string,
) (*JWTClaims, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("chave publica obrigatoria")
	}

	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		},
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(issuer),
		jwt.WithAudience(audience),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, fmt.Errorf("JWT invalido: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("assinatura JWT invalida")
	}

	return claims, nil
}

func (c JWTClaims) HasRole(clientID, role string) bool {
	client, exists := c.ResourceAccess[clientID]
	if !exists {
		return false
	}

	for _, currentRole := range client.Roles {
		if currentRole == role {
			return true
		}
	}

	return false
}
