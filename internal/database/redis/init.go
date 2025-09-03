package redis

import (
	"context"
	"fmt"
	"main/pkg/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
)

func startRedisService() (*redis.Client, error) {
	redisOptions, err := config.LoadRedisConfig()
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(redisOptions)

	if err = testRedisConn(client); err != nil {
		return nil, err
	}

	return client, nil
}

func testRedisConn(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	return nil
}

func InitRedis() {
	service, err := startRedisService()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Redis client: %v", err))
	}

	Client = service
}
