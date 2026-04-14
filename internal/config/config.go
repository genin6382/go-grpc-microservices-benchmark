package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)


type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBURL      string
	JWTSecretKey string
	UserServiceAddr string
	ProductServiceAddr string
	OrderServiceAddr string
	RedisAddr string
	EtcdAddr string
	UserServiceHostAddr string
	ProductServiceHostAddr string
	OrderServiceHostAddr string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	return &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBURL:      fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")),
		JWTSecretKey: os.Getenv("JWT_SECRET_KEY"),
		UserServiceAddr: os.Getenv("USER_SERVICE_GRPC_ADDR"),
		ProductServiceAddr: os.Getenv("PRODUCT_SERVICE_GRPC_ADDR"),
		OrderServiceAddr: os.Getenv("ORDER_SERVICE_GRPC_ADDR"),
		RedisAddr: os.Getenv("REDIS_ADDR"),
		EtcdAddr: os.Getenv("ETCD_ADDR"),
		UserServiceHostAddr: os.Getenv("USER_SERVICE_HOST_ADDR"),
		ProductServiceHostAddr: os.Getenv("PRODUCT_SERVICE_HOST_ADDR"),
		OrderServiceHostAddr: os.Getenv("ORDER_SERVICE_HOST_ADDR"),
	}, nil
}