package solana

import (
	"context"
	"math/big"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func (s *SolClient) GetBalance(address string) (string, error) {
	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return "", err
	}

	out, err := s.Client.GetBalance(
		context.TODO(),
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return "", err
	}

	spew.Dump(out)
	spew.Dump(out.Value)

	lamportsOnAccount := new(big.Float).SetUint64(out.Value)
	solBalance := new(big.Float).Quo(lamportsOnAccount, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))

	return solBalance.Text('f', 9), nil
}
