package middleware

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type AuthMiddleware struct {
	expectedToken string
}

func NewAuthMiddleware(expectedToken string) *AuthMiddleware {
	return &AuthMiddleware{expectedToken: expectedToken}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.expectedToken == "" {
			next(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpx.WriteJson(w, http.StatusUnauthorized, map[string]string{
				"message": "Authorization header missing",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.WriteJson(w, http.StatusUnauthorized, map[string]string{
				"message": "Invalid Authorization format, expected Bearer <token>",
			})
			return
		}

		if parts[1] != m.expectedToken {
			httpx.WriteJson(w, http.StatusUnauthorized, map[string]string{
				"message": "Invalid token",
			})
			return
		}

		next(w, r)
	}
}
