package types

type Network struct {
	Type string `json:"type" enums:"eth_mainnet,eth_ropsten"`
	Name string `json:"name" enums:"ethereum"`
}
