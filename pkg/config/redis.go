package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func LoadRedisConfig() (*redis.Options, error) {
	redisURL := Config.RedisUri
	if redisURL == "" {
		return nil, errors.New("REDIS_URL environment variable not set")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse REDIS_URL: %w", err)
	}

	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second
	opts.PoolSize = 10
	opts.MinIdleConns = 5
	opts.PoolTimeout = 4 * time.Second

	return opts, nil
}

func (c *RedisConfig) ToOptions() *redis.Options {
	return &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password:     c.Password,
		DB:           c.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		PoolTimeout:  4 * time.Second,
	}
}
