package database

import (
	log "ColorStack-Discord-Bot/internal/logger"
	"context"
	"sync"

	_ "github.com/godror/godror"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var redisOnce sync.Once
var redisInstance *RedisClient

// RedisClient manages the connection and operations with Oracle DB.
type RedisClient struct {
	conn *redis.Client
}

// Returns single version of singleton redis
func GetRedisInstance() *RedisClient {
	redisOnce.Do(func() {
		redisClient, err := newRedisClient()
		if err != nil {
			log.Fatal("Failed to initialize redis service: %v", err)
		}
		redisInstance = redisClient
	})

	return redisInstance
}

// NewRedisClient initializes a new DatabaseService instance.
func newRedisClient() (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return &RedisClient{conn: rdb}, nil
}

// Verify if url exists or not
func (rdb *RedisClient) CheckURL(url string) (urlExists bool, err error) {
	result, err := rdb.conn.Exists(ctx, url).Result()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

// Save Job URL
func (rdb *RedisClient) WriteURL(url string) error {
	return rdb.conn.Set(ctx, url, "", 0).Err()
}

// Delete Job URL
func (rdb *RedisClient) DeleteURL(url string) error {
	return rdb.conn.Del(ctx, url).Err()
}
