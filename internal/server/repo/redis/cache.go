package redis

import (
	"context"
	"main/internal/database/redis"
	"main/pkg/models"
	"time"
)

type Cache models.Cache
type CacheService models.CacheImpl

const (
	IpRequestCountPrefix = "ip_request_count:"
	WalletPrefix         = "wallet:"
)

func (c *Cache) GetIpRequestCount(ip string) (int, error) {
	ctx := context.Background()
	key := IpRequestCountPrefix + ip

	val, err := redis.Client.Get(ctx, key).Int()
	if err != nil {
		return 0, err
	}

	return val, nil
}

func (c *Cache) IncrementIpRequestCount(ip string) error {
	ctx := context.Background()
	// will return 0 if key does not exist
	count, _ := c.GetIpRequestCount(ip)
	key := IpRequestCountPrefix + ip
	// set expiry to 10 minutes on first increment
	if count == 0 {
		return redis.Client.Set(ctx, key, 1, 10*time.Minute).Err()
	}
	return redis.Client.Set(ctx, key, count+1, 0).Err()
}

func (c *Cache) SetWallet(wallet, balance string) error {
	ctx := context.Background()
	key := WalletPrefix + wallet
	
	return redis.Client.Set(ctx, key, balance, 10*time.Second).Err()
}

func (c *Cache) GetWallet(wallet string) (string, error) {
	ctx := context.Background()
	key := WalletPrefix + wallet

	val, err := redis.Client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}
