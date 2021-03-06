package main

import (
	"context"
	"math"
	"math/big"

	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const infuraURL = "https://mainnet.infura.io/v3/a6bd0a4b91074054a7d17e27d85d9fac"
const ethAddr = "0x1833be080776553f7c3f83ec9cde2ac216cd5ab9"

func main() {
	conn, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Error(err)
	}
	defer conn.Close()
	account := common.HexToAddress(ethAddr)
	balance, err := conn.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Error(err)
	}
	// convert gwei to eth
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))

	log.Debug("Eth Balance:", ethValue)
}
