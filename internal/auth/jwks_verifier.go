package auth

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func VerifyAccessTokenWithJWKS(
	ctx context.Context,
	rawToken string,
	jwksURL string,
	issuer string,
	audience string,
) (*JWTClaims, error) {
	if rawToken == "" {
		return nil, fmt.Errorf("token obrigatorio")
	}

	// Le o cabecalho sem confiar nele ainda.
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
	)

	token, _, err := parser.ParseUnverified(
		rawToken,
		&JWTClaims{},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"cabecalho JWT invalido: %w", err,
		)
	}

	// Aceita somente RS256.
	if token.Method.Alg() != "RS256" {
		return nil, fmt.Errorf(
			"algoritmo JWT nao permitido",
		)
	}

	kid, ok := token.Header["kid"].(string)
	if !ok || kid == "" {
		return nil, fmt.Errorf("kid obrigatorio")
	}

	// Busca as chaves do Keycloak.
	keys, err := FetchJWKS(ctx, jwksURL)
	if err != nil {
		return nil, err
	}

	// Seleciona a chave publica correta.
	publicKey, err := keys.PublicKeyByKID(kid)
	if err != nil {
		return nil, err
	}

	// Verifica assinatura e claims obrigatorias.
	return VerifySignedAccessToken(
		rawToken,
		publicKey,
		issuer,
		audience,
	)
}
