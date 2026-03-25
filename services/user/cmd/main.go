package main

import (
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	"github.com/genin6382/go-grpc-microservices-benchmark/services/user"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	// Connect to the database
	dbConn, err := db.SetupDatabase(cfg,"internal/db/migrations")
	if err != nil {
        log.Fatalf("Could not setup database: %v", err)
    }
	log.Info("Database migrations applied successfully")
	
	// Set up HTTP server and routes
	router:= chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	userHandler := &user.UserHandler{DB: dbConn, Config: cfg}
	
	//Auth
	router.Post("/login" , userHandler.HandleLogin)
	router.Post("/register", userHandler.HandleCreateUser)

	//Protected routes
	router.Route("/api", func(r chi.Router) {
		r.Use(userHandler.HandleVerifyToken)
		//CRUD
		r.Get("/users", userHandler.HandleListUsers)
		r.Get("/users/{id}", userHandler.HandleGetUserByID)
		r.Delete("/users/{id}", userHandler.HandleDeleteUser)
	})

    log.Info("Server starting on :8080")
    http.ListenAndServe(":8080", router)
}
