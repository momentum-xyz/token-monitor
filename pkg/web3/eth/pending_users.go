package eth

import (
	"context"
	"github.com/momentum-xyz/token-monitor/pkg/cache"
	"github.com/momentum-xyz/token-monitor/pkg/types"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/momentum-xyz/token-monitor/pkg/log"
)

func managePendingUsers(ctx context.Context, id int, client *ethclient.Client, c cache.Cache, contract types.Contract, notify BalanceNotifierFunc, blockCh <-chan uint64, pendingCh <-chan common.Address, activeCh chan<- common.Address) {
	nextBlock := uint64(0)
	users := make([]common.Address, 0)
	ticker := time.NewTicker(time.Second)

	for {
		var err error

		select {
		case <-ctx.Done():
			log.Error(ctx.Err())
			return
		case latestBlock := <-blockCh:
			// the pending users manager queries for the balances at latest block -10
			// the active users manager starts following events from latest block -20
			// in order to not miss any events
			nextBlock = latestBlock - 10
		case user := <-pendingCh:
			users = append(users, user)
		case <-ticker.C:
			if len(users) > 0 && nextBlock > 0 {
				users, err = handlePendingUsers(ctx, id, c, nextBlock, contract, notify, activeCh, users...)
				log.Error(err)
			}
		}
	}
}

func handlePendingUsers(ctx context.Context, id int, c cache.Cache, blockNumber uint64, contract types.Contract, notify BalanceNotifierFunc, activeCh chan<- common.Address, users ...common.Address) ([]common.Address, error) {
	balances := make(map[string]*cache.UserBalance, len(users))
	failed := make(map[string]bool, len(users))
	remainingUsers := make([]common.Address, 0)

	lock := sync.RWMutex{}

	for _, user := range users {
		balances[user.String()] = &cache.UserBalance{
			Balance: big.NewInt(0),
		}
		failed[user.String()] = false
	}

	wg := new(sync.WaitGroup)

	for _, user := range users {
		user := user
		wg.Add(1)

		go func() {
			defer wg.Done()

			// TODO use the multicall contract or the ERC1155 BalanceOfBatch to speed this up
			b, err := contract.BalanceOf(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(blockNumber)}, user)
			if err != nil {
				failed[user.String()] = true
				return
			}

			// Map of a fixed size where each go routine is writing to a different key is thread safe I think? No, added Mutex
			// TODO: instead of mutex, consider using generics with SyncMap for this map and the failed map
			lock.Lock()
			balances[user.String()] = &cache.UserBalance{
				Balance:     b,
				BlockNumber: blockNumber,
			}
			lock.Unlock()
		}()
	}

	wg.Wait() // TODO ctx?

	for user := range balances {
		if failed[user] {
			remainingUsers = append(remainingUsers, common.HexToAddress(user)) // Retry later the users that have no balance
			delete(balances, user)
		}
	}

	err := c.SetRuleTokenBalances(ctx, &cache.TokenBalances{
		RuleID:      id,
		Balances:    balances,
		BlockNumber: blockNumber,
	})
	if err != nil {
		return users, err
	}

	for user, balance := range balances {
		if !failed[user] {
			notify(id, user, balance.Balance)
			activeCh <- common.HexToAddress(user)
		}
	}

	return remainingUsers, nil
}
