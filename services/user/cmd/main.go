package main

import (
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	"github.com/genin6382/go-grpc-microservices-benchmark/services/user"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/db"
	internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
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
	
	// Set up HTTP server and routes
	router:= chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	userHandler := &user.UserHandler{DB: dbConn, Config: cfg}
	
	//Auth
	router.Post("/users/login" , userHandler.HandleLogin)
	router.Post("/users/register", userHandler.HandleCreateUser)

	//Protected routes
	router.Route("/users", func(r chi.Router) {
		r.Use(internalmiddleware.VerifyToken(cfg))
		//CRUD
		r.Get("/", userHandler.HandleListUsers)
		r.Get("/{id}", userHandler.HandleGetUserByID)
		r.Delete("/{id}", userHandler.HandleDeleteUser)
	})

    log.Info("INFO: User-Server starting on :8080")
    http.ListenAndServe(":8080", router)
}
