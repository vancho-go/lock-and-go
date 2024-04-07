package middlewares

import (
	"context"
	"github.com/vancho-go/lock-and-go/internal/service/jwt"
	"net/http"
)

func (m *Middlewares) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		token, err := jwt.GetTokenFromCookie(req)
		if err != nil {
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if validationErr := jwt.IsTokenValid(token); err != nil {
			m.log.Errorf("validation error: %v", validationErr)
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := jwt.GetUserIDFromToken(token)
		if err != nil {
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), jwt.ContextKey, userID)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}
