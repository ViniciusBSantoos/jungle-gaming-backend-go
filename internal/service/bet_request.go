package service

import (
	"errors"
	"strings"
)

type BetRequest struct {
	WalletID              string
	PlayerID              string
	ProviderID            string
	GameID                string
	RoundID               string
	ExternalTransactionID string
	IdempotencyKey        string
	AmountCents           int64
	Currency              string
}

func (r BetRequest) Validate() error {
	if strings.TrimSpace(r.WalletID) == "" ||
		strings.TrimSpace(r.PlayerID) == "" ||
		strings.TrimSpace(r.ProviderID) == "" ||
		strings.TrimSpace(r.GameID) == "" ||
		strings.TrimSpace(r.RoundID) == "" ||
		strings.TrimSpace(r.ExternalTransactionID) == "" ||
		strings.TrimSpace(r.IdempotencyKey) == "" {
		return errors.New("campos obrigatorios ausentes")
	}

	if r.AmountCents <= 0 {
		return errors.New("valor da aposta deve ser positivo")
	}

	if r.Currency != "BRL" {
		return errors.New("moeda nao suportada")
	}

	return nil
}
