package db

import (
	"fmt"
	"database/sql"
	_"github.com/lib/pq" // Postgres driver - Use underscore for blank import to register the driver
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
)

func ConnectToDB(cfg *config.Config) (*sql.DB, error) {

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}