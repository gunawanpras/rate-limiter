package main

import (
	"log"
	"net/http"

	"github.com/gunawanpras/rate-limiter/cache"
	"github.com/gunawanpras/rate-limiter/config"
	ratelimiter "github.com/gunawanpras/rate-limiter/rate-limiter"
)

func main() {
	conf := config.LoadConfig("./config/config.yaml")
	log.Println("Load configuration...")

	rCache := cache.NewRedisCache(conf.Cache)
	log.Println("Starting Redis Cache on port:", conf.Cache.Port)

	rateLimiter := ratelimiter.NewRateLimiter(conf, rCache)
	http.Handle("/rate-limiter", rateLimiter.Handler())

	// Start HTTP server
	log.Println("Starting Rate Limiter on port:", conf.Server.Port)
	if err := http.ListenAndServe(":"+conf.Server.Port, nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
