package cache

import (
	"context"
	"time"
)

type ICache interface {
	SetValue(ctx context.Context, key string, value interface{}, ttl time.Duration) (err error)
	GetValue(ctx context.Context, key string) (string, error)
	DeleteValue(ctx context.Context, key string) error
}
