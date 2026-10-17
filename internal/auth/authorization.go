package auth

import "net/http"

// RequireRole exige uma permissao do cliente informado.
// Deve ser utilizado depois do middleware RequireAuth.
func RequireRole(
	clientID string,
	role string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		principal, authenticated := PrincipalFromContext(
			r.Context(),
		)

		if !authenticated {
			http.Error(
				w,
				"Autenticacao obrigatoria",
				http.StatusUnauthorized,
			)
			return
		}

		if !principal.HasRole(clientID, role) {
			http.Error(
				w,
				"Acesso negado: permissao insuficiente",
				http.StatusForbidden,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
