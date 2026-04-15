package main

import (
	"context"
	"flag"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/genin6382/go-grpc-microservices-benchmark/gateway"
	"github.com/genin6382/go-grpc-microservices-benchmark/gateway/loadbalancer"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	gatewayetcd "github.com/genin6382/go-grpc-microservices-benchmark/internal/etcd"
	internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
)

func main() {
	lbStrategy := flag.String("lb", "round_robin", "Load balancing strategy: round_robin | least_conn | consistent_hash")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var lb loadbalancer.LoadBalancer
	switch *lbStrategy {
	case "round_robin":
		lb = loadbalancer.NewRoundRobin()
	case "least_conn":
		lb = loadbalancer.NewLeastConnections()
	case "consistent_hash":
		lb = loadbalancer.NewConsistentHash(150)
	default:
		log.Fatalf("invalid load balancing strategy: %s", *lbStrategy)
	}

	registry, err := gatewayetcd.NewServiceRegistry([]string{cfg.EtcdAddr})
	if err != nil {
		log.Fatalf("ERROR: Failed to connect to etcd: %v", err)
	}
	defer registry.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := gatewayetcd.WatchService(ctx, registry.Client(), "user", lb); err != nil {
			log.Printf("watch user failed: %v", err)
		}
	}()
	go func() {
		if err := gatewayetcd.WatchService(ctx, registry.Client(), "product", lb); err != nil {
			log.Printf("watch product failed: %v", err)
		}
	}()
	go func() {
		if err := gatewayetcd.WatchService(ctx, registry.Client(), "order", lb); err != nil {
			log.Printf("watch order failed: %v", err)
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("gateway ok"))
	})

	r.Handle("/users/login", gateway.ProxyHandler(lb, "user"))
	r.Handle("/users/register", gateway.ProxyHandler(lb, "user"))

	r.Group(func(r chi.Router) {
		r.Use(internalmiddleware.VerifyToken(cfg))

		r.Handle("/users/*", gateway.WithIdentity(gateway.ProxyHandler(lb, "user")))
		r.Handle("/products/*", gateway.WithIdentity(gateway.ProxyHandler(lb, "product")))
		r.Handle("/orders/*", gateway.WithIdentity(gateway.ProxyHandler(lb, "order")))
	})

	srv := &http.Server{
		Addr:    ":8000",
		Handler: r,
	}

	go func() {
		log.Printf("Gateway running on :8000 with load balancer: %s", *lbStrategy)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("gateway server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gateway...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Println("Gateway stopped")
}