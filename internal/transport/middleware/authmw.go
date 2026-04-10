package middleware

import (
	auth "api-gateway/pkg/api/auth/v1"
	"api-gateway/pkg/authJWT"
	"api-gateway/pkg/response"
	"context"
	"net/http"
	"strings"
)

type ContextKey string

const UserIDKey ContextKey = "user_id"

type authMiddleware struct {
	authService auth.AuthServiceClient
	jwtSecret   string
}

func NewAuthMiddleware(authService auth.AuthServiceClient, jwtSecret string) *authMiddleware {
	return &authMiddleware{
		authService: authService,
		jwtSecret:   jwtSecret}
}

func (am *authMiddleware) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token")
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claim, err := authJWT.ParseToken(token, []byte(am.jwtSecret))
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claim.UserID)
		next(w, r.WithContext(ctx))
	}
}
