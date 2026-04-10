package middleware

import (
	auth "api-gateway/pkg/api/auth/v1"
	"api-gateway/pkg/response"
	"net/http"
	"strings"
)

type authMiddleware struct {
	authService auth.AuthServiceClient
}
func NewAuthMiddleware(authService auth.AuthServiceClient) *authMiddleware {
	return &authMiddleware{authService: authService}
}

func (am *authMiddleware) AuthMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token")
			return
		}
		token := strings.TrimPrefix(header,
	}
}