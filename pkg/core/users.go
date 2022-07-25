package core

import (
	"encoding/json"

	"github.com/OdysseyMomentumExperience/token-service/pkg/web3/eth"
	"github.com/ory/x/errorsx"
)

type NetworkID string

type User struct {
	// Addresses map[NetworkID]*Web3Address
	EthereumAddress string `json:"ethereum_address,omitempty"`
	PolkadotAddress string `json:"polkadot_address,omitempty"`
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
		if user.EthereumAddress != "" {
			_, err := eth.HexToAddress(user.EthereumAddress)
			if err != nil {
				return nil, err
			}
		}
	}
	return users, nil
}
