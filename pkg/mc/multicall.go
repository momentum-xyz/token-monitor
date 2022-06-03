package mc

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/momentum-xyz/token-monitor/pkg/abigen"
	"github.com/ory/x/errorsx"
)

type Call struct {
	Target   common.Address `json:"target"`
	CallData []byte         `json:"call_data"`
}

type CallResponse struct {
	Success    bool   `json:"success"`
	ReturnData []byte `json:"returnData"`
}

func (call Call) GetMultiCall() abigen.Multicall2Call {
	return abigen.Multicall2Call{Target: call.Target, CallData: call.CallData}
}

type EthMultiCaller struct {
	Client          *ethclient.Client
	Abi             abi.ABI
	ContractAddress common.Address
}

func New(client *ethclient.Client) EthMultiCaller {
	// Load Multicall abi for later use
	mcAbi, err := abi.JSON(strings.NewReader(abigen.Multicall2ABI))
	if err != nil {
		panic(err)
	}

	contractAddress := common.HexToAddress("0x5BA1e12693Dc8F9c48aAD8770482f4739bEeD696")

	return EthMultiCaller{
		Client:          client,
		Abi:             mcAbi,
		ContractAddress: contractAddress,
	}
}

func (caller *EthMultiCaller) Execute(calls []Call, blockNumber *big.Int) ([]CallResponse, error) {
	var responses []CallResponse

	var multiCalls = make([]abigen.Multicall2Call, 0, len(calls))

	// Add calls to multicall structure for the contract
	for _, call := range calls {
		multiCalls = append(multiCalls, call.GetMultiCall())
	}

	// Prepare calldata for multicall
	callData, err := caller.Abi.Pack("tryAggregate", false, multiCalls)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	// Perform multicall
	resp, err := caller.Client.CallContract(context.Background(), ethereum.CallMsg{To: &caller.ContractAddress, Data: callData}, blockNumber)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	// Unpack results
	unpackedResp, err := caller.Abi.Unpack("tryAggregate", resp)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	a, err := json.Marshal(unpackedResp[0])
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	err = json.Unmarshal(a, &responses)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	return responses, nil
}
