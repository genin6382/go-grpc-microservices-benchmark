package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	internaljwt "github.com/genin6382/go-grpc-microservices-benchmark/internal/jwt"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func VerifyToken(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
			tokenString = strings.TrimSpace(tokenString)
			if tokenString == "" {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			claims, err := internaljwt.VerifyJWT(tokenString, []byte(cfg.JWTSecretKey))
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireGatewayIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}