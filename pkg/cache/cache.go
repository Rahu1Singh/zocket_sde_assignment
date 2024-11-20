package cache

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// Initialize Redis connection
func InitCache() {
	// Only initialize the Redis client if it's not already initialized
	if redisClient != nil {
		log.Println("Redis client is already initialized.")
		return
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password by default
		DB:       0,                // Default DB
	})

	// Test Redis connection
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully.")
}

// Set key-value pair in the cache
func Set(key string, value string) error {
	// Ensure Redis client is initialized before using it
	if redisClient == nil {
		log.Println("Redis client is not initialized.")
		return nil
	}

	err := redisClient.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		log.Printf("Failed to set cache for key %s: %v", key, err)
		return err
	}
	return nil
}

// Get value from cache by key
func Get(key string) (string, error) {
	// Ensure Redis client is initialized before using it
	if redisClient == nil {
		log.Println("Redis client is not initialized.")
		return "", nil
	}

	val, err := redisClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		// Key does not exist
		return "", nil
	} else if err != nil {
		log.Printf("Failed to get cache for key %s: %v", key, err)
		return "", err
	}
	return val, nil
}

// Delete key from cache
func Delete(key string) error {
	// Ensure Redis client is initialized before using it
	if redisClient == nil {
		log.Println("Redis client is not initialized.")
		return nil
	}

	err := redisClient.Del(context.Background(), key).Err()
	if err != nil {
		log.Printf("Failed to delete cache for key %s: %v", key, err)
		return err
	}
	log.Printf("Cache invalidated for key %s", key)
	return nil
}
