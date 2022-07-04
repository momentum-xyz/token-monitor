package web3

import (
	"encoding/json"
	"github.com/momentum-xyz/token-monitor/pkg/web3/eth"
	"github.com/ory/x/errorsx"
)

type NetworkID string

type User struct {
	// Addresses map[NetworkID]*Web3Address
	EthereumAddress string `json:"ethereum_address"`
	PolkadotAddress string `json:"polkadot_address"`
}

func UnmarshalUser(userJson []byte) (*User, error) {
	var user *User
	err := json.Unmarshal(userJson, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UnmarshalUsers(usersString []byte) ([]*User, error) {
	var users []*User
	err := json.Unmarshal([]byte(usersString), &users)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}
	for _, user := range users {
		_, err := eth.HexToAddress(user.EthereumAddress)
		if err != nil {
			return nil, err
		}
	}
	return users, nil
}
