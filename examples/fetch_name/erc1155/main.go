package main

import (
	"encoding/json"
	"fmt"
	"github.com/OdysseyMomentumExperience/token-service/pkg/abigen"
	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/OdysseyMomentumExperience/token-service/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
)

const infuraURL = "https://mainnet.infura.io/v3/my-rpc"
const contractAddr = "0xA6A5eC7b1B8A34Ff2dcb2926b7c78f52A5ce3b90"
const tokenID = 1724

func main() {
	conn, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Logln(0, err)
	}
	defer conn.Close()

	contract, err := abigen.NewERC1155MetadataURI(common.HexToAddress(contractAddr), conn)
	if err != nil {
		log.Logln(0, "Fatal", err)
	}
	uri, err := contract.Uri(&bind.CallOpts{}, big.NewInt(tokenID))
	uri = strings.ReplaceAll(uri, "ipfs://", "")

	resp, err := http.Get(fmt.Sprintf("https://ipfs.io/ipfs/%s", uri))
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var result types.ERC1155MetaData
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	log.Logln(0, result.Name)
}
