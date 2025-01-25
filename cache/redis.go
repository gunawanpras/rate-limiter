package cache

import (
	"fmt"
	"time"

	rCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

func NewRedisCacheClient(config CacheConfig) (redisCache Cache, err error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		DialTimeout:  time.Duration(config.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
	})

	c := rCache.New(&rCache.Options{
		Redis:      client,
		LocalCache: rCache.NewTinyLFU(1000, time.Minute),
	})

	redisCache = Cache{
		c,
	}

	return

}
