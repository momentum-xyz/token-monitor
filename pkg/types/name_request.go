package types

import "math/big"

type NameRequest struct {
	ContractAddress string   `json:"contractAddress" validate:"required"`
	TokenID         *big.Int `json:"tokenId"`
	TokenType       string   `json:"tokenType" validate:"required" enums:"erc20,erc721,erc1155"`
	Network         Network  `json:"network"`
}
