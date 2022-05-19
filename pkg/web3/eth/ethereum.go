package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

type Client interface {
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
	BlockNumber(ctx context.Context) (uint64, error)
}

func HexToAddress(address string) (common.Address, error) {
	if !common.IsHexAddress(address) {
		return common.Address{}, errors.Errorf("Invalid user address: %s", address)
	}

	return common.HexToAddress(address), nil
}

func HexesToAddresses(addresses []string) ([]common.Address, error) {
	result := make([]common.Address, len(addresses))
	for i, address := range addresses {
		addr, err := HexToAddress(address)
		if err != nil {
			return nil, err
		}
		result[i] = addr
	}

	return result, nil
}
