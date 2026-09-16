package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTokenBucket struct {
	client     *redis.Client
	capacity   float64
	refillRate float64
	keyPrefix  string
}

func NewRedisTokenBucket(client *redis.Client, capacity, refillRate float64, keyPrefix string) *RedisTokenBucket {
	return &RedisTokenBucket{
		client:     client,
		capacity:   capacity,
		refillRate: refillRate,
		keyPrefix:  keyPrefix,
	}
}

func (rtb *RedisTokenBucket) Allow(ctx context.Context, key string) (bool, int64, error) {
	// Lua script for atomic token bucket operations
	// This prevents race conditions between multiple application instances
	script := redis.NewScript(`
        local key = KEYS[1]
        local capacity = tonumber(ARGV[1])
        local refill_rate = tonumber(ARGV[2])
        local now = tonumber(ARGV[3])

        -- Get current bucket state from Redis
        local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
        local tokens = tonumber(bucket[1])
        local last_refill = tonumber(bucket[2])

        -- Initialize bucket if it does not exist
        if tokens == nil then
            tokens = capacity
            last_refill = now
        end

        -- Calculate tokens to add based on elapsed time
        local elapsed = now - last_refill
        tokens = math.min(capacity, tokens + (elapsed * refill_rate))

        local allowed = 0
        if tokens >= 1 then
            tokens = tokens - 1
            allowed = 1
        end

        -- Update bucket state in Redis with expiry to auto-cleanup inactive keys
        redis.call('HSET', key, 'tokens', tokens, 'last_refill', now)
        redis.call('EXPIRE', key, 3600)  -- Expire after 1 hour of inactivity

        return {allowed, tokens}
    `)

	redisKey := fmt.Sprintf("%s:%s", rtb.keyPrefix, key)
	now := float64(time.Now().Unix())

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	res, err := script.Run(ctxWithTimeout, rtb.client, []string{redisKey}, rtb.capacity, rtb.refillRate, now).Slice()

	if err != nil {
		return false, 0, fmt.Errorf("redis script execution failed: %w", err)
	}
	allowed := res[0].(int64) == 1
	remaining := res[1].(int64)
	return allowed, remaining, nil
}

func (rtb *RedisTokenBucket) GetTokens(ctx context.Context, key string) (float64, error) {
	redisKey := fmt.Sprintf("%s:%s", rtb.keyPrefix, key)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	res, err := rtb.client.HGet(ctxWithTimeout, redisKey, "tokens").Result()

	if err == redis.Nil {
		return rtb.capacity, nil
	}
	if err != nil {
		return 0, err
	}

	tokens, err := strconv.ParseFloat(res, 64)
	if err != nil {
		return 0, err
	}
	return tokens, nil
}
