package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	internaljwt "github.com/genin6382/go-grpc-microservices-benchmark/internal/jwt"
)

func VerifyToken(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if tokenString == "" {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			claims, err := internaljwt.VerifyJWT(tokenString, []byte(cfg.JWTSecretKey))
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
