package auth

import "fmt"

type Config struct {
	Issuer   string
	Audience string
}

func NewConfig(issuer, audience string) (Config, error) {
	if issuer == "" {
		return Config{}, fmt.Errorf("issuer obrigatorio")
	}

	if audience == "" {
		return Config{}, fmt.Errorf("audience obrigatoria")
	}

	return Config{
		Issuer:   issuer,
		Audience: audience,
	}, nil
}
