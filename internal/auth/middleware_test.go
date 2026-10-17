package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestRequireAuth(t *testing.T) {
	validator := func(
		ctx context.Context,
		token string,
	) (*JWTClaims, error) {
		if token != "token-valido" {
			return nil, errors.New("token invalido")
		}

		return &JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "jogador-123",
			},
		}, nil
	}

	protected := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		principal, ok := PrincipalFromContext(
			r.Context(),
		)

		if !ok || principal.Subject != "jogador-123" {
			t.Error("Usuario autenticado nao encontrado")
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := RequireAuth(validator, protected)

	tests := []struct {
		name           string
		authorization  string
		expectedStatus int
	}{
		{"SemToken", "", http.StatusUnauthorized},
		{"TokenInvalido", "Bearer errado", http.StatusUnauthorized},
		{"TokenValido", "Bearer token-valido", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/protected",
				nil,
			)

			if tt.authorization != "" {
				req.Header.Set(
					"Authorization",
					tt.authorization,
				)
			}

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf(
					"Esperado HTTP %d, recebido %d",
					tt.expectedStatus,
					recorder.Code,
				)
			}
		})
	}
}
