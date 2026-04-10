package middleware

import (
	"api-gateway/pkg/response"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token")
			return
		}
		token := strings.TrimPrefix(header,
	}
}