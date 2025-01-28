package ratelimiter

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gunawanpras/rate-limiter/cache"
	"github.com/gunawanpras/rate-limiter/config"
	"github.com/gunawanpras/rate-limiter/helper"
)

type RateLimiter struct {
	config config.Config
	cache  cache.ICache
	mutex  sync.Mutex
}

type Visitor struct {
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
// error: An error is returned when there is an errors during the operation.
func (rl *RateLimiter) Allow(ip string) (bool, error) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	ctx := context.Background()
	v, err := rl.cache.GetValue(ctx, ip)
	if err != nil {
		v = "{}"
		if !strings.Contains(err.Error(), "key is missing") {
			return false, err
		}
	}

	var storedVisitor Visitor

	errUnmarshal := json.Unmarshal([]byte(v), &storedVisitor)
	if errUnmarshal != nil {
		return false, errUnmarshal
	}

	if v == "{}" || helper.TimeHelper.GetElapsedTime(storedVisitor.LastSeen) > time.Duration(rl.config.RateLimiter.Interval)*time.Second {
		storedVisitor = Visitor{
			LastSeen: helper.TimeHelper.Now(),
			Count:    1,
		}

		storedVisitorBytes, err := json.Marshal(storedVisitor)
		if err != nil {
			return false, err
		}

		if err = rl.cache.SetValue(ctx, ip, storedVisitorBytes, time.Duration(rl.config.Cache.Ttl)*time.Minute); err != nil {
			return false, err
		}

		return true, nil
	}

	if storedVisitor.Count < rl.config.RateLimiter.Limit {
		storedVisitor.Count += 1
		storedVisitor.LastSeen = helper.TimeHelper.Now()

		storedVisitorBytes, err := json.Marshal(storedVisitor)
		if err != nil {
			return false, err
		}

		if err = rl.cache.SetValue(ctx, ip, storedVisitorBytes, time.Duration(rl.config.Cache.Ttl)*time.Minute); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (rl *RateLimiter) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := helper.ReqHelper.GetIp(r)

		allow, err := rl.Allow(ip)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !allow {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("Request allowed"))
	})
}
