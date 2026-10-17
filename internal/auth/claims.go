package auth

import (
	"encoding/json"
	"fmt"
)

type ClientRoles struct {
	Roles []string `json:"roles"`
}

type AccessClaims struct {
	Subject        string                 `json:"sub"`
	TokenType      string                 `json:"typ"`
	PreferredUser  string                 `json:"preferred_username"`
	ResourceAccess map[string]ClientRoles `json:"resource_access"`
}

func ParseAccessClaims(data []byte) (AccessClaims, error) {
	var claims AccessClaims

	if err := json.Unmarshal(data, &claims); err != nil {
		return AccessClaims{}, fmt.Errorf(
			"erro ao ler claims: %w", err,
		)
	}

	if claims.Subject == "" {
		return AccessClaims{}, fmt.Errorf(
			"subject obrigatorio",
		)
	}

	return claims, nil
}

func (c AccessClaims) HasRole(
	clientID string,
	role string,
) bool {
	client, exists := c.ResourceAccess[clientID]
	if !exists {
		return false
	}

	for _, r := range client.Roles {
		if r == role {
			return true
		}
	}

	return false
}

func (c AccessClaims) ValidateTokenType() error {
	if c.TokenType != "Bearer" {
		return fmt.Errorf(
			"tipo de token invalido: esperado Bearer",
		)
	}

	return nil
}
