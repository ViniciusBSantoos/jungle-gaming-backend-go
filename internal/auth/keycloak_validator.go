package auth

import (
	"context"
	"fmt"
	"strings"
)

func NewKeycloakValidator(
	issuer string,
	audience string,
) (TokenValidator, error) {
	issuer = strings.TrimRight(issuer, "/")

	if issuer == "" || audience == "" {
		return nil, fmt.Errorf(
			"issuer e audience sao obrigatorios",
		)
	}

	jwksURL := issuer +
		"/protocol/openid-connect/certs"

	validate := func(
		ctx context.Context,
		token string,
	) (*JWTClaims, error) {
		return VerifyAccessTokenWithJWKS(
			ctx,
			token,
			jwksURL,
			issuer,
			audience,
		)
	}

	return validate, nil
}
