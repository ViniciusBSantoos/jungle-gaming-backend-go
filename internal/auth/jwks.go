package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func FetchJWKS(
	ctx context.Context,
	jwksURL string,
) (JWKS, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		jwksURL,
		nil,
	)
	if err != nil {
		return JWKS{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return JWKS{}, fmt.Errorf(
			"erro ao buscar chaves: %w", err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return JWKS{}, fmt.Errorf(
			"Keycloak retornou HTTP %d",
			resp.StatusCode,
		)
	}

	var keys JWKS

	err = json.NewDecoder(
		http.MaxBytesReader(nil, resp.Body, 1<<20),
	).Decode(&keys)

	if err != nil {
		return JWKS{}, fmt.Errorf(
			"erro ao ler chaves: %w", err,
		)
	}

	if len(keys.Keys) == 0 {
		return JWKS{}, fmt.Errorf(
			"nenhuma chave encontrada",
		)
	}

	return keys, nil
}

func (j JWK) PublicKey() (*rsa.PublicKey, error) {
	if j.Kty != "RSA" {
		return nil, fmt.Errorf("tipo de chave nao suportado")
	}

	if j.N == "" || j.E == "" {
		return nil, fmt.Errorf("chave RSA incompleta")
	}

	modulus, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, fmt.Errorf("modulo RSA invalido: %w", err)
	}

	exponent, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, fmt.Errorf("expoente RSA invalido: %w", err)
	}

	n := new(big.Int).SetBytes(modulus)
	e := new(big.Int).SetBytes(exponent)

	if n.Sign() <= 0 || !e.IsInt64() {
		return nil, fmt.Errorf("parametros RSA invalidos")
	}

	exponentValue := e.Int64()

	if exponentValue < 3 ||
		exponentValue > int64(^uint(0)>>1) ||
		exponentValue%2 == 0 {
		return nil, fmt.Errorf("expoente RSA invalido")
	}

	return &rsa.PublicKey{
		N: n,
		E: int(exponentValue),
	}, nil
}

func (j JWKS) PublicKeyByKID(
	kid string,
) (*rsa.PublicKey, error) {
	if kid == "" {
		return nil, fmt.Errorf("kid obrigatorio")
	}

	for _, key := range j.Keys {
		if key.Kid != kid {
			continue
		}

		if key.Kty != "RSA" ||
			key.Use != "sig" ||
			key.Alg != "RS256" {
			return nil, fmt.Errorf(
				"chave nao autorizada para RS256",
			)
		}

		publicKey, err := key.PublicKey()
		if err != nil {
			return nil, err
		}

		if publicKey.N.BitLen() < 2048 {
			return nil, fmt.Errorf(
				"chave RSA menor que 2048 bits",
			)
		}

		return publicKey, nil
	}

	return nil, fmt.Errorf(
		"chave publica nao encontrada",
	)
}
