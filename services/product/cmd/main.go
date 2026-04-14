package main

import (
	"net/http"
	"net"
	"google.golang.org/grpc"
	"time"
	"os/signal"
	"context"
	"syscall"

	pb "github.com/genin6382/go-grpc-microservices-benchmark/pb/product"

	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/db"
	"github.com/genin6382/go-grpc-microservices-benchmark/services/product"
	internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/cache"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/etcd"


	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
)

func main (){
	//Load config
	cfg , err := config.LoadConfig()
	if err != nil {
		log.Fatalf("ERROR: Failed to Load Configurations: %v",err)
	}
	//Connect to DB
	dbConn , err := db.SetupDatabase(cfg, "internal/db/migrations")
	if err != nil {
		log.Fatalf("ERROR: Failed to Setup Database: %v",err)
	}
	defer dbConn.Close()
	log.Info("INFO: Successfully Setup Database!")


	//Setup gRPC server
	go func(){
		lis , err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("ERROR: Failed to Listen on port 50052: %v",err)
		}
		grpcServer := grpc.NewServer()
		pb.RegisterProductServiceServer(grpcServer, &product.Server{DB: dbConn})
		log.Info("INFO: gRPC Server running on :50052")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("ERROR: Failed to Serve gRPC Server: %v",err)
		}
	}()	

	// Setup Redis Cache
	redisClient := cache.SetupRedisCache(cfg.RedisAddr)

	if redisClient == nil {
		log.Warn("WARN: Redis cache is not available. Proceeding without cache.")
	} else {
		log.Info("INFO: Redis cache is set up and ready to use.")
	}

	//Setup Chi router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	productHandler := &product.ProductHandler{DB:dbConn, Config:cfg, CacheClient: redisClient}
	//Protected Routes
	router.Route("/products",func(r chi.Router){
		r.Use(internalmiddleware.VerifyToken(cfg))
		//GET
		r.Get("/",productHandler.HandleListProducts)
		r.Get("/{id}",productHandler.HandleGetProductByID)
		//CREATE
		r.Post("/",productHandler.HandleCreateProduct)
		//UPDATE
		r.Put("/{id}",productHandler.HandleUpdateProductDetails)
		r.Patch("/{id}/stock",productHandler.HandleUpdateProductStock)
		//DELETE
		r.Delete("/{id}",productHandler.HandleDeleteProduct)
	})

	log.Info("INFO: Product-Server starting on :8081")

    // etcd registration
    etcdClient, err := etcd.NewServiceRegistry([]string{cfg.EtcdAddr})
    if err != nil {
        log.Fatalf("ERROR: Failed to connect to etcd: %v", err)
    }

    // Graceful shutdown context — declare BEFORE using ctx anywhere
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Start HTTP server to constantly 
    srv := &http.Server{Addr: ":8081", Handler: router}
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Register in etcd — blocks here until SIGTERM
    // Lease expires automatically on ctx cancel = automatic deregistration
    if err := etcdClient.Register(ctx, "product", cfg.ProductServiceHostAddr, 10); err != nil {
    log.Fatalf("ERROR: etcd registration error: %v", err)
	}

	<-ctx.Done()

	log.Info("Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("Server stopped")

}