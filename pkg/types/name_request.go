package types

import "math/big"

type NameRequest struct {
	ContractAddress string   `json:"contractAddress" validate:"required"`
	TokenID         *big.Int `json:"tokenId"`
	TokenType       string   `json:"tokenType" validate:"required"`
	Network         Network  `json:"network"`
}
