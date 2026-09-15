package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(client *redis.Client) *RedisClient {
	return &RedisClient{
		client: client,
	}
}

func (rc *RedisClient) Set(ctx context.Context, key string, value any, duration time.Duration) error {
	valueJson, err := json.Marshal(value)
	if err != nil {
		return err
	}

	cacheTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return rc.client.Set(cacheTimeout, key, valueJson, duration).Err()
}

func (rc *RedisClient) Get(ctx context.Context, key string, dest any) error {
	cacheTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	cached, err := rc.client.Get(cacheTimeout, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(cached), dest)
}

func (rc *RedisClient) Delete(ctx context.Context, key string) error {
	cacheTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return rc.client.Del(cacheTimeout, key).Err()
}

func (rc *RedisClient) Increment(ctx context.Context, key string) error {
	cacheTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return rc.client.Incr(cacheTimeout, key).Err()
}
