package eth

import (
	"github.com/OdysseyMomentumExperience/token-service/pkg/constants"
	"github.com/OdysseyMomentumExperience/token-service/pkg/networks"
	"github.com/OdysseyMomentumExperience/token-service/pkg/types"
	"github.com/pkg/errors"
)

type EthNameService struct {
	contract types.Contract
	clients  *networks.ClientManager
}

func NewEthNameService(clients *networks.ClientManager) *EthNameService {
	return &EthNameService{
		clients: clients,
	}
}

func (s *EthNameService) GetTokenName(req *types.NameRequest) (string, error) {
	tokenType := req.TokenType
	network, err := s.clients.GetEthereumClient(req.Network.Name)
	if err != nil {
		return "", err
	}
	switch tokenType {
	case constants.TokenTypeERC20:
		s.contract, err = NewERC20Contract(req.ContractAddress, network)
	case constants.TokenTypeERC721:
		s.contract, err = NewERC721Contract(req.ContractAddress, network)
	case constants.TokenTypeERC1155:
		if req.TokenID == nil {
			return "", errors.Errorf("tokenId needs to be provided for erc1155 tokens")
		}
		s.contract, err = NewERC1155Contract(req.ContractAddress, req.TokenID, network)
	default:
		return "", errors.Errorf("unsupported token type %s", tokenType)
	}
	if err != nil {
		return "", err
	}

	tokenName, err := s.contract.TokenName(nil)
	if err != nil {
		return "", err
	}
	return tokenName, nil
}
