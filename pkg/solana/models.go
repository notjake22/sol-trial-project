package solana

import (
	"main/pkg/config"

	"github.com/gagliardetto/solana-go/rpc"
)

type SolClient struct {
	Client *rpc.Client
}

func NewSolClient() *SolClient {
	return &SolClient{
		Client: rpc.New(config.Config.RpcUri),
	}
}
