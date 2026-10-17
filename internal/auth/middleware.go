package auth

import (
	"context"
	"net/http"
	"strings"
)

type principalContextKey struct{}

type TokenValidator func(
	ctx context.Context,
	token string,
) (*JWTClaims, error)

func RequireAuth(
	validate TokenValidator,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		authorization := r.Header.Get("Authorization")

		parts := strings.Fields(authorization)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			parts[1] == "" {
			http.Error(
				w,
				"Token Bearer obrigatorio",
				http.StatusUnauthorized,
			)
			return
		}

		if validate == nil {
			http.Error(
				w,
				"Autenticacao indisponivel",
				http.StatusServiceUnavailable,
			)
			return
		}

		claims, err := validate(
			r.Context(),
			parts[1],
		)

		if err != nil || claims == nil ||
			claims.Subject == "" {
			http.Error(
				w,
				"Token invalido",
				http.StatusUnauthorized,
			)
			return
		}

		principal := NewPrincipal(claims)

		ctx := context.WithValue(
			r.Context(),
			principalContextKey{},
			principal,
		)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}

func PrincipalFromContext(
	ctx context.Context,
) (Principal, bool) {
	principal, ok := ctx.Value(
		principalContextKey{},
	).(Principal)

	return principal, ok
}
