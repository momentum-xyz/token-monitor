package eth

import (
	"math/big"

	"github.com/OdysseyMomentumExperience/token-service/pkg/abigen"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type erc721Contract struct {
	contract *abigen.ERC721
}

func newERC721Contract(addressHex string, client bind.ContractBackend) (*erc721Contract, error) {
	contractAddress, err := HexToAddress(addressHex)
	if err != nil {
		return nil, err
	}
	contract, err := abigen.NewERC721(contractAddress, client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ERC721 contract for %s", contractAddress.String())
	}
	return &erc721Contract{
		contract: contract,
	}, nil
}

func (t *erc721Contract) getLogs(opts *bind.FilterOpts, userAddresses []common.Address) ([]logItem, error) {
	res := make([]logItem, 0)
	senderLogs, err := t.contract.FilterTransfer(opts, userAddresses, nil, nil)
	if err != nil {
		return nil, err
	}
	for senderLogs.Next() {
		res = append(res, logItem{
			user:  senderLogs.Event.From,
			value: big.NewInt(-1),
		})
	}
	receiverLogs, err := t.contract.FilterTransfer(opts, nil, userAddresses, nil)
	if err != nil {
		return nil, err
	}
	for receiverLogs.Next() {
		res = append(res, logItem{
			user:  receiverLogs.Event.To,
			value: big.NewInt(1),
		})
	}
	return res, nil
}

func (t *erc721Contract) balanceOf(opts *bind.CallOpts, userAddress common.Address) (*big.Int, error) {
	return t.contract.BalanceOf(opts, userAddress)
}
