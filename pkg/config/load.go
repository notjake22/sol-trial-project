package config

import "os"

func Load() {
	Config = &Structure{
		RpcUri: os.Getenv("RPC_URI"),
	}
}
