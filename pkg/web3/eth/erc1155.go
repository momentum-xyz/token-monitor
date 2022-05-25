package eth

import (
	"encoding/json"
	"fmt"
	"github.com/OdysseyMomentumExperience/token-service/pkg/constants"
	"github.com/OdysseyMomentumExperience/token-service/pkg/types"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"github.com/OdysseyMomentumExperience/token-service/pkg/abigen"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type erc1155Contract struct {
	tokenID  *big.Int
	contract *abigen.ERC1155
	metadata *abigen.ERC1155MetadataURI
}

func newERC1155Contract(addressHex string, tokenID *big.Int, client bind.ContractBackend) (*erc1155Contract, error) {
	contractAddress, err := HexToAddress(addressHex)
	if err != nil {
		return nil, err
	}
	contract, err := abigen.NewERC1155(contractAddress, client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create erc1155 contract for %s", contractAddress.String())
	}
	metadata, err := abigen.NewERC1155MetadataURI(contractAddress, client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create erc1155 metadata contract for %s", contractAddress.String())
	}
	return &erc1155Contract{
		contract: contract,
		metadata: metadata,
	}, nil
}

func (t *erc1155Contract) getLogs(opts *bind.FilterOpts, userAddresses []common.Address) ([]logItem, error) {
	res := make([]logItem, 0)
	senderLogs, err := t.contract.FilterTransferSingle(opts, nil, userAddresses, nil)
	if err != nil {
		return nil, err
	}
	for senderLogs.Next() {
		res = append(res, logItem{
			user:  senderLogs.Event.From,
			value: new(big.Int).Neg(senderLogs.Event.Value),
		})
	}
	receiverLogs, err := t.contract.FilterTransferSingle(opts, nil, nil, userAddresses)
	if err != nil {
		return nil, err
	}
	for receiverLogs.Next() {
		res = append(res, logItem{
			user:  receiverLogs.Event.To,
			value: receiverLogs.Event.Value,
		})
	}

	senderBatchLogs, err := t.contract.FilterTransferBatch(opts, nil, userAddresses, nil)
	if err != nil {
		return nil, err
	}
	for senderBatchLogs.Next() {
		for i, id := range senderBatchLogs.Event.Ids {
			if id.Cmp(t.tokenID) == 0 {
				res = append(res, logItem{
					user:  senderBatchLogs.Event.From,
					value: new(big.Int).Neg(senderBatchLogs.Event.Values[i]),
				})
			}
		}
	}
	receiverBatchLogs, err := t.contract.FilterTransferBatch(opts, nil, nil, userAddresses)
	if err != nil {
		return nil, err
	}
	for receiverBatchLogs.Next() {
		for i, id := range receiverBatchLogs.Event.Ids {
			if id.Cmp(t.tokenID) == 0 {
				res = append(res, logItem{
					user:  receiverBatchLogs.Event.From,
					value: receiverBatchLogs.Event.Values[i],
				})
			}
		}
	}
	return res, nil
}

func (t *erc1155Contract) balanceOf(opts *bind.CallOpts, userAddress common.Address) (*big.Int, error) {
	return t.contract.BalanceOf(opts, userAddress, t.tokenID)
}

func (t *erc1155Contract) tokenName(opts *bind.CallOpts) (string, error) {
	uri, err := t.metadata.Uri(opts, t.tokenID)
	if err != nil {
		return "", err
	}
	uri = strings.ReplaceAll(uri, constants.IPFSPrefix, "")
	resp, err := http.Get(fmt.Sprintf("%s/%s", constants.IPFSURLPrefix, uri))
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var result types.ERC1155MetaData
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	return result.Name, nil
}
