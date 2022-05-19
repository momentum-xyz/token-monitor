package web3

import (
	"context"
	"errors"
	"github.com/OdysseyMomentumExperience/token-service/pkg/cache"
	"math/big"

	"github.com/OdysseyMomentumExperience/token-service/pkg/constants"
	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/OdysseyMomentumExperience/token-service/pkg/networks"
	"github.com/OdysseyMomentumExperience/token-service/pkg/web3/eth"
)

type TokenTracker interface {
	Stop(ctx context.Context)
	AddUserAddress(user string) error
}

type RuleStatePublisher interface {
	PublishRuleState(id int, user string, satisfied bool)
}

// NOTE: not safe for concurrent use
type RuleManager struct {
	users    []*User
	trackers map[int]TokenTracker

	clients   *networks.ClientManager
	publisher RuleStatePublisher
	cache     cache.Cache
}

func NewRuleManager(clients *networks.ClientManager, publisher RuleStatePublisher, cache cache.Cache) *RuleManager {
	lm := &RuleManager{
		publisher: publisher,
		clients:   clients,
		trackers:  make(map[int]TokenTracker),
		cache:     cache,
	}

	return lm
}

func (lm *RuleManager) Init(ctx context.Context, rules []*RuleDefinition, users []*User) error {
	lm.users = users
	for _, rule := range rules {
		err := lm.StartNewTokenTracker(ctx, rule)
		if err != nil {
			return err
		}
	}
	return nil
}

// called sequentially
func (lm *RuleManager) StartNewTokenTracker(ctx context.Context, rule *RuleDefinition) error {
	var err error
	switch lm.clients.GetNetworkType(rule.Network) {
	case constants.NetworkTypeEthereum:
		err = lm.StartNewEthereumTokenTracker(ctx, rule)
	// case constants.NetworkTypePolkadot:
	// 	err = lm.StartNewPolkadotTokenTracker(ctx, rule)
	default:
		return errors.New("unsupported network type")
	}
	if err != nil {
		return err
	}
	return nil
}

func (lm *RuleManager) StartNewEthereumTokenTracker(ctx context.Context, rule *RuleDefinition) error {
	l, ok := lm.trackers[rule.ID]
	if ok {
		l.Stop(ctx)
		delete(lm.trackers, rule.ID)
	}

	var err error
	ethclient, err := lm.clients.GetEthereumClient(rule.Network)
	if err != nil {
		return err
	}

	// copy the users list in order to avoid concurrent access
	users := getEthereumAddresses(lm.users)

	l, err = eth.StartNewBalanceTracker(ctx, rule.ID, rule.Token.Type, rule.Token.TokenID, ethclient, rule.Token.Address, users, lm.cache, notifier(rule, lm.publisher))

	if err != nil {
		return err
	}

	lm.trackers[rule.ID] = l

	return nil
}

func notifier(rule *RuleDefinition, publisher RuleStatePublisher) func(id int, user string, balance *big.Int) {
	return func(id int, user string, balance *big.Int) {
		go publisher.PublishRuleState(id, user, balance.Cmp(rule.Requirements.MinBalance) >= 0)
	}
}

func getEthereumAddresses(users []*User) []string {
	addresses := make([]string, len(users))
	for i, u := range users {
		if u.EthereumAddress != "" {
			addresses[i] = u.EthereumAddress
		}
	}
	return addresses
}

func (lm *RuleManager) AddUser(user *User) error {
	lm.users = append(lm.users, user)

	for _, l := range lm.trackers {
		err := l.AddUserAddress(user.EthereumAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

func (lm *RuleManager) DeleteRule(ctx context.Context, id int) {
	l, ok := lm.trackers[id]
	if ok {
		l.Stop(ctx)
		delete(lm.trackers, id)
	} else {
		log.Logln(0, "rule manager", "could not find rule with id:", id)
	}
}
