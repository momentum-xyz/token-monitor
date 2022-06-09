package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Contract interface {
	GetLogs(opts *bind.FilterOpts, userAddresses []common.Address) ([]LogItem, error)
	BalanceOf(opts *bind.CallOpts, userAddress common.Address) (*big.Int, error)
	TokenName(opts *bind.CallOpts) (string, error)
}

type LogItem struct {
	User  common.Address
	Value *big.Int
}
