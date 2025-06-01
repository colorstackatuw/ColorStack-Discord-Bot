package database

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func createTestRedisClient(addr string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisClient{conn: rdb}
}

func TestRedisClient_WriteCheckDeleteURL(t *testing.T) {
	// Start a fake in-memory Redis server
	mockRedis, err := miniredis.Run()
	assert.NoError(t, err)
	defer mockRedis.Close()

	rdb := createTestRedisClient(mockRedis.Addr())

	testURL := "test-url-key"

	// Ensure the key does not exist
	exists, err := rdb.CheckURL(testURL)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Write the key
	err = rdb.WriteURL(testURL)
	assert.NoError(t, err)

	// Now the key should exist
	exists, err = rdb.CheckURL(testURL)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Delete the key
	err = rdb.DeleteURL(testURL)
	assert.NoError(t, err)

	// Key should be gone
	exists, err = rdb.CheckURL(testURL)
	assert.NoError(t, err)
	assert.False(t, exists)
}
