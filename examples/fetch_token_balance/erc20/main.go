package main

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/momentum-xyz/token-monitor/pkg/abigen"
	"github.com/momentum-xyz/token-monitor/pkg/log"
)

const infuraURL = "https://ropsten.infura.io/v3/a6bd0a4b91074054a7d17e27d85d9fac"

// const contractAddr = "0x655f2166b0709cd575202630952d71e2bb0d61af"
// const userAddr = "0x501e71ec141e031d804c48fbfc1c0a5b020c827f"

const contractAddr = "0xc778417E063141139Fce010982780140Aa0cD5Ab"

// const contractAddr = "0x655f2166b0709cd575202630952d71e2bb0d61af"
// const contractAddr = "0x110a13fc3efe6a245b50102d2d79b3e76125ae83"

const userAddr = "0xa428d6424f9430b293eccc49fa71971b8a9d258d"

func main() {
	conn, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Logln(0, err)
	}
	defer conn.Close()
	//ERC20
	contract, err := abigen.NewERC20(common.HexToAddress(contractAddr), conn)
	if err != nil {
		log.Logln(0, "Fatal", err)
	}

	bn, err := conn.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("Block Number:", bn)
	amt, err := contract.BalanceOf(&bind.CallOpts{
		BlockNumber: big.NewInt(int64(bn - 10)),
	}, common.HexToAddress(userAddr))

	if err != nil {
		panic(err)
	}
	fBalance := new(big.Float)
	fBalance.SetString(amt.String())
	bal := new(big.Float).Quo(fBalance, big.NewFloat(math.Pow10(18)))
	log.Logln(0, "ERC 20 Balance:", bal)
}
