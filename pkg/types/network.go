package types

type Network struct {
	Type string `json:"type" enums:"eth_mainnet,eth_ropsten" validate:"required"`
	Name string `json:"name" enums:"ethereum" validate:"required"`
}
