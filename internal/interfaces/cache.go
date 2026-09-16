package interfaces

import (
	"context"
	"time"
)

type CacheInterface interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	Delete(ctx context.Context, key string) error
	Increment(ctx context.Context, key string) error
}
