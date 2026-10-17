package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type Verifier struct {
	verifier *oidc.IDTokenVerifier
}

func NewVerifier(
	ctx context.Context,
	config Config,
) (*Verifier, error) {
	provider, err := oidc.NewProvider(
		ctx,
		config.Issuer,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"erro ao conectar com Keycloak: %w",
			err,
		)
	}

	tokenVerifier := provider.Verifier(
		&oidc.Config{
			ClientID: config.Audience,
		},
	)

	return &Verifier{
		verifier: tokenVerifier,
	}, nil
}

func (v *Verifier) Verify(
	ctx context.Context,
	rawToken string,
) (*oidc.IDToken, error) {
	token, err := v.verifier.Verify(
		ctx,
		rawToken,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"token invalido: %w",
			err,
		)
	}

	return token, nil
}

func (v *Verifier) VerifyAccessToken(
	ctx context.Context,
	rawToken string,
) (AccessClaims, error) {
	if strings.TrimSpace(rawToken) == "" {
		return AccessClaims{}, fmt.Errorf(
			"token de acesso obrigatorio",
		)
	}

	token, err := v.Verify(ctx, rawToken)
	if err != nil {
		return AccessClaims{}, err
	}

	var claims AccessClaims

	if err := token.Claims(&claims); err != nil {
		return AccessClaims{}, fmt.Errorf(
			"erro ao interpretar token: %w", err,
		)
	}

	if claims.Subject == "" {
		return AccessClaims{}, fmt.Errorf(
			"token sem identificacao do usuario",
		)
	}
	if err := claims.ValidateTokenType(); err != nil {
		return AccessClaims{}, err
	}
	return claims, nil
}
