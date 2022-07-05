package main

import (
	"github.com/OdysseyMomentumExperience/token-service/pkg/abigen"
	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const infuraURL = "https://goerli.infura.io/v3/a6bd0a4b91074054a7d17e27d85d9fac"
const contractAddr = "0x655f2166b0709cd575202630952d71e2bb0d61af"
const userAddr = "0x501E71EC141e031D804c48fBFC1C0a5b020C827F"

func main() {
	conn, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Error(err)
	}
	defer conn.Close()

	//ERC721
	contract, err := abigen.NewERC721(common.HexToAddress(contractAddr), conn)
	if err != nil {
		log.Error(err)
	}
	amt, _ := contract.BalanceOf(&bind.CallOpts{}, common.HexToAddress(userAddr))
	log.Debug("ERC 721 Balance:", amt)
}
