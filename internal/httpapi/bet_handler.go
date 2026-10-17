package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"jungle-gaming-backend-go/internal/service"
)

type CreateBetRequest struct {
	ProviderID            string `json:"provider_id"`
	GameID                string `json:"game_id"`
	RoundID               string `json:"round_id"`
	ExternalTransactionID string `json:"external_transaction_id"`
	IdempotencyKey        string `json:"idempotency_key"`
	AmountCents           int64  `json:"amount_cents"`
}

func (r CreateBetRequest) ToServiceRequest(
	walletID string,
	playerID string,
) (service.BetRequest, error) {
	req := service.BetRequest{
		WalletID:              walletID,
		PlayerID:              playerID,
		ProviderID:            r.ProviderID,
		GameID:                r.GameID,
		RoundID:               r.RoundID,
		ExternalTransactionID: r.ExternalTransactionID,
		IdempotencyKey:        r.IdempotencyKey,
		AmountCents:           r.AmountCents,
		Currency:              "BRL",
	}

	if err := req.Validate(); err != nil {
		return service.BetRequest{}, fmt.Errorf(
			"aposta invalida: %w",
			err,
		)
	}

	return req, nil
}

func DecodeCreateBetRequest(
	r *http.Request,
) (CreateBetRequest, error) {
	var request CreateBetRequest

	decoder := json.NewDecoder(
		http.MaxBytesReader(nil, r.Body, 1<<20),
	)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		return CreateBetRequest{}, fmt.Errorf(
			"JSON invalido: %w",
			err,
		)
	}

	return request, nil
}
