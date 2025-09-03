package models

type Cache struct {
	TTLDefaultSeconds int
}

type CacheImpl interface {
	GetIpRequestCount(ip string) (int, error)
	IncrementIpRequestCount(ip string) error
	SetWallet(wallet, balance string) error
	GetWallet(wallet string) (string, error)
}
