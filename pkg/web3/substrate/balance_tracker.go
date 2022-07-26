package substrate

import (
	"context"
	"math/big"

	"github.com/OdysseyMomentumExperience/token-service/pkg/cache"
	"github.com/OdysseyMomentumExperience/token-service/pkg/core"
	"github.com/OdysseyMomentumExperience/token-service/pkg/networks"
	"github.com/OdysseyMomentumExperience/token-service/pkg/redis"
	"github.com/ethereum/go-ethereum/common"
)

type BalanceTracker struct {
	id            int
	cache         *redis.Cache
	cancel        context.CancelFunc
	client        *networks.SubstrateClient
	notify        BalanceNotifierFunc
	pendingUserCh chan common.Address
}

type BalanceNotifierFunc func(id int, user string, balance *big.Int)

func StartNewBalanceTracker(ctx context.Context,
	rule *core.RuleDefinition,
	client *networks.SubstrateClient,
	users []*core.User,
	cache cache.Cache,
	notify BalanceNotifierFunc) (*BalanceTracker, error) {

	return nil, nil
}

func (l *BalanceTracker) AddUserAddress(addressHex string) error {

	return nil
}

func (l *BalanceTracker) Stop(ctx context.Context) {
	l.cache.ClearRuleTokenBalances(ctx, l.id)
	l.cancel()
}

func getPolkadotAddresses(users []*core.User) []string {
	addresses := make([]string, len(users))
	for i, u := range users {
		if u.PolkadotAddress != "" {
			addresses[i] = u.PolkadotAddress
		}
	}
	return addresses
}
