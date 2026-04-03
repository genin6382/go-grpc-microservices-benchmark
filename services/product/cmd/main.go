package main

import (
	"net/http"
	"net"
	"google.golang.org/grpc"

	pb "github.com/genin6382/go-grpc-microservices-benchmark/pb/product"

	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/db"
	"github.com/genin6382/go-grpc-microservices-benchmark/services/product"
	internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"

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

	//Setup Chi router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	productHandler := &product.ProductHandler{DB:dbConn, Config:cfg}
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
    http.ListenAndServe(":8081", router)
}