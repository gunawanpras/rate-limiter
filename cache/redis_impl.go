package cache

import (
	"context"
	"log"

	"time"

	rCache "github.com/go-redis/cache/v8"
)

type RedisImpl struct {
	cache *Cache
}

func NewRedisCache(config CacheConfig) ICache {
	c, err := NewRedisCacheClient(config)
	if err != nil {
		log.Panic(err)
	}

	return &RedisImpl{
		cache: &c,
	}
}

func (r *RedisImpl) SetValue(ctx context.Context, key string, value interface{}, ttl time.Duration) (err error) {
	err = r.cache.Set(&rCache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
		TTL:   ttl,
	})

	if err != nil {
		return
	}

	return
}

func (r *RedisImpl) GetValue(ctx context.Context, key string) (value string, err error) {
	if err = r.cache.Get(ctx, key, &value); err != nil {
		return
	}

	return
}

func (r *RedisImpl) DeleteValue(ctx context.Context, key string) (err error) {
	if err = r.cache.Delete(ctx, key); err != nil {
		return
	}

	return
}
