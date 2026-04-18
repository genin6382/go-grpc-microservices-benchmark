package main

import (
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	"github.com/genin6382/go-grpc-microservices-benchmark/services/user"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/db"
	internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/etcd"

	"time"
	"os/signal"
	"context"
	"syscall"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net"
	"google.golang.org/grpc"
	pb "github.com/genin6382/go-grpc-microservices-benchmark/pb/user"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/cache"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("ERROR: Failed to load config: %v", err)
	}
	// Connect to the database
	dbConn, err := db.SetupDatabase(cfg,"internal/db/migrations")
	if err != nil {
        log.Fatalf("ERROR: Could not setup database: %v", err)
    }
	defer dbConn.Close()
	log.Info("INFO: Database migrations applied successfully")

	// gRPC server setup 

	go func() {
        lis, err := net.Listen("tcp", ":50051")
        if err != nil {
            log.Fatalf("gRPC failed to listen: %v", err)
        }
        grpcServer := grpc.NewServer()
        pb.RegisterUserServiceServer(grpcServer, &user.Server{DB: dbConn}) 
        log.Info("gRPC server running on :50051")
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("gRPC failed to serve: %v", err)
        }
    }()

	// Set up Redis cache
	cacheClient := cache.SetupRedisCache(cfg.RedisAddr)

	if cacheClient == nil {
		log.Warn("WARNING: Redis cache not available, proceeding without caching")
	} else {
		log.Info("INFO: Redis cache connected successfully")
	}

	// Set up HTTP server and routes
	router:= chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	userHandler := &user.UserHandler{DB: dbConn, Config: cfg, CacheClient: cacheClient}
	
	//Auth
	router.Post("/users/login" , userHandler.HandleLogin)
	router.Post("/users/register", userHandler.HandleCreateUser)

	//Protected routes
	router.Route("/users", func(r chi.Router) {
		r.Use(internalmiddleware.RequireGatewayIdentity)
		//CRUD
		r.Get("/", userHandler.HandleListUsers)
		r.Get("/{id}", userHandler.HandleGetUserByID)
		r.Delete("/{id}", userHandler.HandleDeleteUser)
	})

    log.Info("INFO: User-Server starting on :8080")
    // etcd registration
    etcdClient, err := etcd.NewServiceRegistry([]string{cfg.EtcdAddr})
    if err != nil {
        log.Fatalf("ERROR: Failed to connect to etcd: %v", err)
    }

    // Graceful shutdown context — declare BEFORE using ctx anywhere
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Start HTTP server to constantly 
    srv := &http.Server{Addr: ":8080", Handler: router}
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Register in etcd — blocks here until SIGTERM
    // Lease expires automatically on ctx cancel = automatic deregistration
    if err := etcdClient.Register(ctx, "user", cfg.UserServiceHostAddr, 10); err != nil {
    log.Fatalf("ERROR: etcd registration error: %v", err)
	}

	<-ctx.Done()

	log.Info("Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("Server stopped")
}
