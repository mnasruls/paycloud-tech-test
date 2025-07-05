package main

import (
	"coding_test_2/internal/config"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

// connectRedis establishes a connection to Redis
func connectRedis(ctx context.Context) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: config.REDIS_URL,
		DB:   0, // use default DB
	})

	// Use the provided context for the Ping operation
	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis:", config.REDIS_URL)
	return rdb, nil
}
