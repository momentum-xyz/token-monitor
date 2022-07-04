package main

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/momentum-xyz/token-monitor/pkg/abigen"
	"github.com/momentum-xyz/token-monitor/pkg/log"
)

const infuraURL = "https://mainnet.infura.io/v3/my-rpc"
const contractAddr = "0x655f2166b0709cd575202630952d71e2bb0d61af"
const userAddr = "0x501E71EC141e031D804c48fBFC1C0a5b020C827F"

func main() {
	conn, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Logln(0, err)
	}
	defer conn.Close()

	//ERC721
	contract, err := abigen.NewERC721(common.HexToAddress(contractAddr), conn)
	if err != nil {
		log.Logln(0, "Fatal", err)
	}
	amt, _ := contract.BalanceOf(&bind.CallOpts{}, common.HexToAddress(userAddr))
	log.Logln(0, "ERC 721 Balance:", amt)
}
