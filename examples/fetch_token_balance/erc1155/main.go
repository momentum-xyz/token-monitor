package main

import (
	"math/big"

	"github.com/OdysseyMomentumExperience/token-service/pkg/abigen"
	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const infuraURL = "https://goerli.infura.io/v3/a6bd0a4b91074054a7d17e27d85d9fac"
const contractAddr = "0x2e3ef7931f2d0e4a7da3dea950ff3f19269d9063"
const userAddr = "0x501E71EC141e031D804c48fBFC1C0a5b020C827F"
const tokenID = 123

func main() {
	conn, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Error(err)
	}
	defer conn.Close()

	//ERC1155 ?.
	contract, err := abigen.NewERC1155(common.HexToAddress(contractAddr), conn)
	if err != nil {
		log.Error(err)
	}
	amt, _ := contract.BalanceOf(&bind.CallOpts{}, common.HexToAddress(userAddr), big.NewInt(tokenID))
	log.Debug("ERC 1155 Balance:", amt)
}
