package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
    "github.com/genin6382/go-grpc-microservices-benchmark/internal/db"
    internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
    "github.com/genin6382/go-grpc-microservices-benchmark/services/order"
    productpb "github.com/genin6382/go-grpc-microservices-benchmark/pb/product"
    userpb    "github.com/genin6382/go-grpc-microservices-benchmark/pb/user"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    log "github.com/sirupsen/logrus"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("ERROR: Failed to load config: %v", err)
    }

    dbConn, err := db.SetupDatabase(cfg, "internal/db/migrations")
    if err != nil {
        log.Fatalf("ERROR: Failed to setup database: %v", err)
    }
    defer dbConn.Close()
    log.Info("INFO: Database setup successful")

    // gRPC client — User Service
    userConn, err := grpc.NewClient(cfg.UserServiceAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("ERROR: Failed to connect to user-service: %v", err)
    }
    defer userConn.Close()

    // gRPC client — Product Service
    productConn, err := grpc.NewClient(cfg.ProductServiceAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("ERROR: Failed to connect to product-service: %v", err)
    }
    defer productConn.Close()

    // Wire dependencies
    userGrpcClient    := order.NewUserServiceClient(userpb.NewUserServiceClient(userConn))
    productGrpcClient := order.NewProductServiceClient(productpb.NewProductServiceClient(productConn))

    orderHandler := &order.OrderHandler{
        DB:            dbConn,
        Config:        cfg,
        UserClient:    userGrpcClient,
        ProductClient: productGrpcClient,
    }

    // Router
    router := chi.NewRouter()
    router.Use(middleware.Logger)
    router.Use(middleware.Recoverer)

    router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok"}`))
    })

    router.Route("/orders", func(r chi.Router) {
        r.Use(internalmiddleware.VerifyToken(cfg))
        r.Get("/", orderHandler.HandleListOrders)
        r.Get("/{id}", orderHandler.HandleGetOrderByID)
        r.Get("/user/{user_id}", orderHandler.HandleGetOrdersByUserID)
        r.Post("/", orderHandler.HandleCreateOrder)
        r.Patch("/{id}", orderHandler.HandleUpdateOrderStatus)
        r.Delete("/{id}", orderHandler.HandleDeleteOrder)
    })

    // Graceful shutdown
    srv := &http.Server{Addr: ":8082", Handler: router}
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        log.Info("INFO: Order-Server starting on :8082")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    <-quit
    log.Info("Shutting down gracefully...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
    log.Info("Server stopped")
}