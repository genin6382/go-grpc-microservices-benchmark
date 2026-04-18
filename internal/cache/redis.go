package cache

import (
	"context"
	"time"
	
	 "github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

func SetupRedisCache(address string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		PoolSize: 10,
		DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
		MaxRetries: 4,
	})
	
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Debugf("DEBUG: Failed to connect to Redis: %v\n", err)
		return nil
	}
	
	return rdb
}