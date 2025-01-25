package cache

import "github.com/go-redis/cache/v8"

type CacheConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	Db           int    `yaml:"db"`
	Ttl          int    `yaml:"ttl"`
	DialTimeout  int    `yaml:"dial_timeout"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

type Cache struct {
	*cache.Cache
}
