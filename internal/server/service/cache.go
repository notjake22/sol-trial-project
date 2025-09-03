package service

import "main/internal/server/repo/redis"

var (
	cache        = &redis.Cache{}
	cacheService = redis.CacheService(cache)
)

func GetIpRequestCount(ip string) (int, error) {
	res, err := cacheService.GetIpRequestCount(ip)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func IncrementIpRequestCount(ip string) error {
	err := cacheService.IncrementIpRequestCount(ip)
	if err != nil {
		return err
	}

	return nil
}

func SetWallet(wallet, balance string) error {
	err := cacheService.SetWallet(wallet, balance)
	if err != nil {
		return err
	}

	return nil
}

func GetWallet(wallet string) (string, error) {
	res, err := cacheService.GetWallet(wallet)
	if err != nil {
		return "", err
	}

	return res, nil
}
