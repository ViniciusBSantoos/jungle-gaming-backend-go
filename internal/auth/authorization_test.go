package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRole(t *testing.T) {
	const clientID = "jungle-backend-api"

	tests := []struct {
		name           string
		authenticated  bool
		roles          []string
		expectedStatus int
	}{
		{
			name:           "SemAutenticacao",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "SemPermissao",
			authenticated:  true,
			roles:          []string{"wallet:read"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "ComPermissao",
			authenticated:  true,
			roles:          []string{"wallet:write"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protected := http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(http.StatusOK)
			})

			handler := RequireRole(
				clientID,
				"wallet:write",
				protected,
			)

			req := httptest.NewRequest(
				http.MethodPost,
				"/wallet",
				nil,
			)

			if tt.authenticated {
				principal := Principal{
					Subject: "jogador-123",
					Roles: map[string]ClientRoles{
						clientID: {
							Roles: tt.roles,
						},
					},
				}

				ctx := context.WithValue(
					req.Context(),
					principalContextKey{},
					principal,
				)

				req = req.WithContext(ctx)
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
