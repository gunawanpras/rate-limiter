package config

import (
	"log"
	"os"

	"github.com/gunawanpras/rate-limiter/cache"
	"gopkg.in/yaml.v2"
)

type (
	Server struct {
		Port string `yaml:"port"`
	}

	RateLimiter struct {
		Limit    int `yaml:"limit"`
		Interval int `yaml:"interval"`
	}

	Config struct {
		Server      Server            `yaml:"server"`
		RateLimiter RateLimiter       `yaml:"rate_limiter"`
		Cache       cache.CacheConfig `yaml:"cache"`
	}
)

func LoadConfig(filePath string) Config {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Error opening config file:", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatal("Error decoding config file:", err)
	}

	return config
}
