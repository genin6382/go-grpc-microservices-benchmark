package db

import (
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)
func SetupDatabase(cfg *config.Config, migrationPath string) (*sql.DB, error){
	dbConn, err := ConnectToDB(cfg)
	if err != nil {
		return nil, err
	}
	m, err := migrate.New(
		"file://"+migrationPath,
		cfg.DBURL,
	)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	return dbConn, nil
}