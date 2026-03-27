package main

import (
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/db"
	internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
	"github.com/genin6382/go-grpc-microservices-benchmark/services/order"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"net/http"
	log "github.com/sirupsen/logrus"
)


func main (){
	cfg , err := config.LoadConfig()

	if err!= nil {
		log.Fatalf("ERROR: Failed to Load Configurations: %v",err)
	}

	//Connect to DB
	dbConn , err := db.SetupDatabase(cfg, "internal/db/migrations")
	if err != nil {
		log.Fatalf("ERROR: Failed to Setup Database: %v",err)
	}
	defer dbConn.Close()
	log.Info("INFO: Successfully Setup Database!")

	//Setup Chi router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	orderHandler := &order.OrderHandler{DB:dbConn, Config:cfg}

	//Protected Routes
	router.Route("/orders",func(r chi.Router){
		r.Use(internalmiddleware.VerifyToken(cfg))

		r.Get("/", orderHandler.HandleListOrders)
		r.Get("/{id}", orderHandler.HandleGetOrderByID)
		r.Get("/user/{user_id}", orderHandler.HandleGetOrdersByUserID)

		r.Post("/", orderHandler.HandleCreateOrder)

		r.Patch("/{id}", orderHandler.HandleUpdateOrderStatus)
		
		r.Delete("/{id}", orderHandler.HandleDeleteOrder)
	})

	log.Info("INFO: Product-Server starting on :8082")
    http.ListenAndServe(":8082", router)
}