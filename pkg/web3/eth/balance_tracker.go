package eth

import (
	"context"
	"fmt"
	"github.com/OdysseyMomentumExperience/token-service/pkg/cache"
	"math/big"
	"time"

	"github.com/OdysseyMomentumExperience/token-service/pkg/constants"
	"github.com/OdysseyMomentumExperience/token-service/pkg/redis"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

type BalanceTracker struct {
	id            int
	cache         *redis.Cache
	cancel        context.CancelFunc
	client        *ethclient.Client
	notify        BalanceNotifierFunc
	contract      contract
	pendingUserCh chan common.Address
}

type Cache interface {
}

type BalanceNotifierFunc func(id int, user string, balance *big.Int)

func StartNewBalanceTracker(ctx context.Context, id int, tokenType string, tokenID *big.Int, client *ethclient.Client, contractAddressHex string, users []string, cache cache.Cache, notify BalanceNotifierFunc) (*BalanceTracker, error) {
	var contract contract
	var err error

	switch tokenType {
	case constants.TokenTypeERC20:
		contract, err = newERC20Contract(contractAddressHex, client)
	case constants.TokenTypeERC721:
		contract, err = newERC721Contract(contractAddressHex, client)
	case constants.TokenTypeERC1155:
		contract, err = newERC1155Contract(contractAddressHex, tokenID, client)
	default:
		return nil, errors.Errorf("unsupported token type %s", tokenType)
	}
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	pendingUserCh := make(chan common.Address, 100)

	l := &BalanceTracker{
		cancel:        cancel,
		notify:        notify,
		client:        client,
		contract:      contract,
		pendingUserCh: pendingUserCh,
	}

	go initialize(ctx, id, client, users, pendingUserCh, cache, contract, notify)

	return l, nil
}

type contract interface {
	getLogs(opts *bind.FilterOpts, userAddresses []common.Address) ([]logItem, error)
	balanceOf(opts *bind.CallOpts, userAddress common.Address) (*big.Int, error)
	tokenName(opts *bind.CallOpts) (string, error)
}

func initialize(ctx context.Context, id int, client *ethclient.Client, users []string, pendingUserCh chan common.Address, c cache.Cache, contract contract, notify BalanceNotifierFunc) {
	// init:
	// get latest block number
	// for each user:
	//    if we have a balance in the cache for the current user
	//         send the user to the active users channel
	//    else
	//         send the user to the pending users channel

	// pending user manager:
	// 	 when a new user arrives - check their balance by querying the blockchain
	// 	 send the user to the active users channel
	// active user manager:
	// init:
	//    follow the events starting at the block number we have in the cache
	//    ignore the events for the users that are synced to a later block until it's reached

	blockCh := make(chan uint64)
	blockCh2 := make(chan uint64)
	go startBlockChecker(ctx, id, client, blockCh, blockCh2)

	activeUserCh := make(chan common.Address, 100)
	addresses, _ := HexesToAddresses(users)
	ub := getCachedBalancesWithRetry(ctx, id, c, addresses)

	if ub != nil { // we have some cached balances, so we don't have to start from the beginning
		for _, user := range addresses {
			if ub[user.String()] != nil {
				activeUserCh <- user
			} else {
				pendingUserCh <- user
			}
		}
	} else {
		for _, user := range addresses {
			pendingUserCh <- user
		}
	}

	blockNumber := <-blockCh
	fmt.Println(blockNumber)
	go managePendingUsers(ctx, id, client, c, contract, notify, blockCh, pendingUserCh, activeUserCh)
	go manageActiveUsers(ctx, id, c, contract, notify, client, blockNumber, activeUserCh, blockCh)
}

func getCachedBalancesWithRetry(ctx context.Context, id int, c cache.Cache, addresses []common.Address) map[string]*cache.UserBalance {
	ticker := time.NewTicker(time.Second * 5)
	var tb map[string]*cache.UserBalance
	var err error

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			tb, err = c.GetRuleTokenBalance(ctx, id)
			if err == nil {
				return tb
			}
		}
	}
}

func (l *BalanceTracker) AddUserAddress(addressHex string) error {
	user, err := HexToAddress(addressHex)
	if err != nil {
		return err
	}

	l.pendingUserCh <- user

	return nil
}

var sleepTime = 10 * time.Second

func (l *BalanceTracker) Stop(ctx context.Context) {
	l.cache.ClearRuleTokenBalances(ctx, l.id)
	l.cancel()
}
