package ratelimiter

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gunawanpras/rate-limiter/cache"
	"github.com/gunawanpras/rate-limiter/config"
)

type RateLimiter struct {
	config config.Config
	cache  cache.ICache
	mutex  sync.Mutex
}

type visitor struct {
	LastSeen time.Time
	Count    int
}

func NewRateLimiter(config config.Config, rCache cache.ICache) *RateLimiter {
	return &RateLimiter{
		config: config,
		cache:  rCache,
	}
}

// Allow checks if a request from a given IP address is allowed based on the rate limit configuration.
// It retrieves the visitor's data from the cache, updates the request count and timestamp, and stores
// it back in the cache. If the visitor is new or their last request was beyond the allowed interval,
// start the counter from the beginning.
//
// Parameters:
//
// ip: The IP address of the incoming request.
//
// Returns:
//
// bool: True if the request is allowed, false if the rate limit is exceeded.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	ctx := context.Background()
	v, err := rl.cache.GetValue(ctx, ip)

	var storedVisitor visitor
	_ = json.Unmarshal([]byte(v), &storedVisitor)

	if err != nil || time.Since(storedVisitor.LastSeen) > time.Duration(rl.config.RateLimiter.Interval)*time.Second {
		storedVisitor = visitor{
			LastSeen: time.Now(),
			Count:    1,
		}

		storedVisitorBytes, err := json.Marshal(storedVisitor)
		if err != nil {
			return false
		}

		if err = rl.cache.SetValue(ctx, ip, storedVisitorBytes, time.Duration(rl.config.Cache.Ttl)*time.Minute); err != nil {
			return false
		}

		return true
	}

	if storedVisitor.Count < rl.config.RateLimiter.Limit {
		storedVisitor.Count += 1
		storedVisitor.LastSeen = time.Now()

		storedVisitorBytes, err := json.Marshal(storedVisitor)
		if err != nil {
			return false
		}

		if err = rl.cache.SetValue(ctx, ip, storedVisitorBytes, time.Duration(rl.config.Cache.Ttl)*time.Minute); err != nil {
			return false
		}

		return true
	}

	return false
}

func (rl *RateLimiter) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !rl.Allow(ip) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("Request allowed"))
	})
}
