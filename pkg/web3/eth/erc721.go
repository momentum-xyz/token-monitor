package eth

import (
	"github.com/OdysseyMomentumExperience/token-service/pkg/types"
	"math/big"

	"github.com/OdysseyMomentumExperience/token-service/pkg/abigen"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type erc721Contract struct {
	contract *abigen.ERC721
	metadata *abigen.ERC721Metadata
}

func NewERC721Contract(addressHex string, client bind.ContractBackend) (*erc721Contract, error) {
	contractAddress, err := HexToAddress(addressHex)
	if err != nil {
		return nil, err
	}
	contract, err := abigen.NewERC721(contractAddress, client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ERC721 contract for %s", contractAddress.String())
	}
	metadata, err := abigen.NewERC721Metadata(contractAddress, client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ERC721 metadata for %s", contractAddress.String())
	}
	return &erc721Contract{
		contract: contract,
		metadata: metadata,
	}, nil
}

func (t *erc721Contract) GetLogs(opts *bind.FilterOpts, userAddresses []common.Address) ([]types.LogItem, error) {
	res := make([]types.LogItem, 0)
	senderLogs, err := t.contract.FilterTransfer(opts, userAddresses, nil, nil)
	if err != nil {
		return nil, err
	}
	for senderLogs.Next() {
		res = append(res, types.LogItem{
			User:  senderLogs.Event.From,
			Value: big.NewInt(-1),
		})
	}
	receiverLogs, err := t.contract.FilterTransfer(opts, nil, userAddresses, nil)
	if err != nil {
		return nil, err
	}
	for receiverLogs.Next() {
		res = append(res, types.LogItem{
			User:  receiverLogs.Event.To,
			Value: big.NewInt(1),
		})
	}
	return res, nil
}

func (t *erc721Contract) BalanceOf(opts *bind.CallOpts, userAddress common.Address) (*big.Int, error) {
	return t.contract.BalanceOf(opts, userAddress)
}

func (t *erc721Contract) TokenName(opts *bind.CallOpts) (string, error) {
	return t.metadata.Name(opts)
}
