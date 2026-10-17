package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func (r BetRequest) PayloadHash() (string, error) {
	payload := struct {
		WalletID              string `json:"wallet_id"`
		PlayerID              string `json:"player_id"`
		ProviderID            string `json:"provider_id"`
		GameID                string `json:"game_id"`
		RoundID               string `json:"round_id"`
		ExternalTransactionID string `json:"external_transaction_id"`
		AmountCents           int64  `json:"amount_cents"`
		Currency              string `json:"currency"`
	}{
		WalletID:              r.WalletID,
		PlayerID:              r.PlayerID,
		ProviderID:            r.ProviderID,
		GameID:                r.GameID,
		RoundID:               r.RoundID,
		ExternalTransactionID: r.ExternalTransactionID,
		AmountCents:           r.AmountCents,
		Currency:              r.Currency,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
