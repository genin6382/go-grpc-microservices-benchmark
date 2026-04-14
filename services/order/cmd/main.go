package main

import (
    "context"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
    "github.com/genin6382/go-grpc-microservices-benchmark/internal/db"
    internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
    "github.com/genin6382/go-grpc-microservices-benchmark/internal/etcd" 
    productpb "github.com/genin6382/go-grpc-microservices-benchmark/pb/product"
    userpb "github.com/genin6382/go-grpc-microservices-benchmark/pb/user"
    "github.com/genin6382/go-grpc-microservices-benchmark/services/order"
    "github.com/genin6382/go-grpc-microservices-benchmark/internal/cache"

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

    // gRPC clients
    userConn, err := grpc.NewClient(cfg.UserServiceAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("ERROR: Failed to connect to user-service: %v", err)
    }
    defer userConn.Close()

    productConn, err := grpc.NewClient(cfg.ProductServiceAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("ERROR: Failed to connect to product-service: %v", err)
    }
    defer productConn.Close()

    userGrpcClient    := order.NewUserServiceClient(userpb.NewUserServiceClient(userConn))
    productGrpcClient := order.NewProductServiceClient(productpb.NewProductServiceClient(productConn))

    //setup redis cache
    cacheClient := cache.SetupRedisCache(cfg.RedisAddr)
    if cacheClient == nil {
        log.Warn("WARNING: Redis cache not available, proceeding without caching")
    } else {
        log.Info("INFO: Redis cache connected successfully")
    }

    orderHandler := &order.OrderHandler{
        DB:            dbConn,
        Config:        cfg,
        UserClient:    userGrpcClient,
        ProductClient: productGrpcClient,
        CacheClient:   cacheClient,
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

    // etcd registration
    etcdClient, err := etcd.NewServiceRegistry([]string{cfg.EtcdAddr})
    if err != nil {
        log.Fatalf("ERROR: Failed to connect to etcd: %v", err)
    }

    // Graceful shutdown context — declare BEFORE using ctx anywhere
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Start HTTP server to constantly 
    srv := &http.Server{Addr: ":8082", Handler: router}
    go func() {
        log.Info("INFO: Order service starting on :8082")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Register in etcd — blocks here until SIGTERM
    // Lease expires automatically on ctx cancel = automatic deregistration
    if err := etcdClient.Register(ctx, "order", cfg.OrderServiceHostAddr, 10); err != nil {
    log.Fatalf("ERROR: etcd registration error: %v", err)
	}

	<-ctx.Done()

	log.Info("Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("Server stopped")
}