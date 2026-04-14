package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/genin6382/go-grpc-microservices-benchmark/gateway"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("gateway ok"))
	})

	userProxy := gateway.NewReverseProxy("http://localhost:8080")
	productProxy := gateway.NewReverseProxy("http://localhost:8081")
	orderProxy := gateway.NewReverseProxy("http://localhost:8082")

	r.Handle("/users/login", userProxy)
	r.Handle("/users/register", userProxy)

	r.Group(func(r chi.Router) {
		r.Use(internalmiddleware.VerifyToken(cfg))

		r.Handle("/users/*", gateway.WithIdentity(userProxy))
        r.Handle("/products/*", gateway.WithIdentity(productProxy))
        r.Handle("/orders/*", gateway.WithIdentity(orderProxy))
	})

	log.Println("Gateway running on :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
