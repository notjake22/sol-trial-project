package models

type WalletsRequest struct {
	Wallets []string `json:"wallets"`
}

type WalletBalance struct {
	Wallet  string `json:"wallet"`
	Balance string `json:"balance"`
	Cache   string `json:"cache"`
}
