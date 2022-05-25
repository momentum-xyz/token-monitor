package eth

import (
	"math/big"

	"github.com/OdysseyMomentumExperience/token-service/pkg/abigen"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type erc20Contract struct {
	contract *abigen.ERC20
}

func newERC20Contract(addressHex string, client bind.ContractBackend) (*erc20Contract, error) {
	contractAddress, err := HexToAddress(addressHex)
	if err != nil {
		return nil, err
	}
	contract, err := abigen.NewERC20(contractAddress, client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ERC20 contract for %s", contractAddress.String())
	}
	return &erc20Contract{
		contract: contract,
	}, nil
}

func (t *erc20Contract) getLogs(opts *bind.FilterOpts, userAddresses []common.Address) ([]logItem, error) {
	res := make([]logItem, 0)
	senderLogs, err := t.contract.FilterTransfer(opts, userAddresses, nil)
	if err != nil {
		return nil, err
	}
	for senderLogs.Next() {
		res = append(res, logItem{
			user:  senderLogs.Event.From,
			value: new(big.Int).Neg(senderLogs.Event.Value),
		})
	}
	receiverLogs, err := t.contract.FilterTransfer(opts, nil, userAddresses)
	if err != nil {
		return nil, err
	}
	for receiverLogs.Next() {
		res = append(res, logItem{
			user:  receiverLogs.Event.To,
			value: receiverLogs.Event.Value,
		})
	}
	return res, nil
}

func (t *erc20Contract) balanceOf(opts *bind.CallOpts, userAddress common.Address) (*big.Int, error) {
	return t.contract.BalanceOf(opts, userAddress)
}

func (t *erc20Contract) tokenName(opts *bind.CallOpts) (string, error) {
	return t.contract.Name(opts)
}
