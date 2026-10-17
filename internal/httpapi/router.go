package httpapi

import (
	"encoding/json"
	"errors"
	"jungle-gaming-backend-go/internal/auth"
	"jungle-gaming-backend-go/internal/repository"
	"jungle-gaming-backend-go/internal/service"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func NewRouter(
	validate auth.TokenValidator,
	wallets *repository.WalletRepository,
	bets *service.BetService,
) http.Handler {

	mux := http.NewServeMux()

	// Rota publica: nao exige autenticacao.
	mux.HandleFunc("GET /health", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	// Rota protegida: exige JWT e permissao wallet:read.

	walletHandler := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			http.Error(w, "Usuario nao autenticado", http.StatusUnauthorized)
			return
		}

		if wallets == nil {
			http.Error(w, "Servico de carteira indisponivel", http.StatusServiceUnavailable)
			return
		}

		wallet, err := wallets.FindByPlayerID(
			r.Context(),
			principal.Subject,
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "Carteira nao encontrada", http.StatusNotFound)
				return
			}

			http.Error(w, "Erro ao consultar carteira", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"wallet_id":     wallet.ID,
			"player_id":     wallet.PlayerID,
			"currency":      wallet.Currency,
			"balance_cents": wallet.BalanceCents,
		})
	})

	mux.Handle(
		"GET /wallet/me",
		auth.RequireAuth(
			validate,
			auth.RequireRole(
				"jungle-backend-api",
				"wallet:read",
				walletHandler,
			),
		),
	)

	mux.Handle(
		"POST /bets",
		auth.RequireAuth(
			validate,
			auth.RequireRole(
				"jungle-backend-api",
				"wallet:write",
				http.HandlerFunc(func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					if wallets == nil || bets == nil {
						http.Error(w, "Servico indisponivel", http.StatusServiceUnavailable)
						return
					}

					principal, ok := auth.PrincipalFromContext(r.Context())
					if !ok {
						http.Error(w, "Usuario nao autenticado", http.StatusUnauthorized)
						return
					}

					input, err := DecodeCreateBetRequest(r)
					if err != nil {
						http.Error(w, "JSON invalido", http.StatusBadRequest)
						return
					}

					wallet, err := wallets.FindByPlayerID(
						r.Context(),
						principal.Subject,
					)
					if err != nil {
						if errors.Is(err, pgx.ErrNoRows) {
							http.Error(w, "Carteira nao encontrada", http.StatusNotFound)
							return
						}

						http.Error(w, "Erro ao consultar carteira", http.StatusInternalServerError)
						return
					}

					req, err := input.ToServiceRequest(
						wallet.ID,
						principal.Subject,
					)
					if err != nil {
						http.Error(w, "Dados da aposta invalidos", http.StatusBadRequest)
						return
					}

					result, err := bets.ProcessBet(r.Context(), req)
					if err != nil {
						http.Error(w, "Nao foi possivel processar a aposta", http.StatusConflict)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)

					json.NewEncoder(w).Encode(map[string]interface{}{
						"transaction_id": result.TransactionID,
						"status":         result.Status,
						"balance_cents":  result.BalanceCents,
						"replayed":       result.Replayed,
					})
				}),
			),
		),
	)

	return mux
}
