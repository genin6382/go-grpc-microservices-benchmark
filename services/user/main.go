package main

import (
    log "github.com/sirupsen/logrus"
    "github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
    "github.com/genin6382/go-grpc-microservices-benchmark/internal/db"
)

func main() {
	cfg , err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConn, err := db.ConnectToDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()
	
	log.Info("Successfully connected to the database")
}
