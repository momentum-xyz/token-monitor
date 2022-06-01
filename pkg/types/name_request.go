package types

import "math/big"

type NameRequest struct {
	ContractAddress string   `json:"contractAddress"`
	TokenID         *big.Int `json:"tokenId"`
	TokenType       string   `json:"tokenType"`
	Network         Network  `json:"network"`
}
