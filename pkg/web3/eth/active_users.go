package eth

import (
	"context"
	"math/big"
	"time"

	"github.com/OdysseyMomentumExperience/token-service/pkg/cache"
	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/OdysseyMomentumExperience/token-service/pkg/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func manageActiveUsers(ctx context.Context, id int, c cache.Cache, contract types.Contract, notify BalanceNotifierFunc, client Client, lastCheckedBlock uint64, userCh chan common.Address, blockCh chan uint64) {
	ticker := time.NewTicker(time.Second * 3)

	activeUsers := make([]common.Address, 0)
	var latestBlock uint64

	for {
		log.Debug("rule:", id, "-", "user list:", activeUsers)

		var err error

		select {
		case <-ctx.Done():
			log.Debug("rule:", id, "-", "Stopping active user manager", ctx.Err())
			return
		case user := <-userCh:
			activeUsers = append(activeUsers, user)
			// new active user whose balance we just checked, so no need to verify again on the current block
			continue
		case blockNumber := <-blockCh:
			if blockNumber >= 20 {
				latestBlock = blockNumber - 20
			} else {
				latestBlock = 0
			}
		case <-ticker.C:
		}

		if latestBlock > 0 && latestBlock > lastCheckedBlock {
			lastCheckedBlock, err = handleNetworkPoll(ctx, id, c, contract, notify, lastCheckedBlock, latestBlock, client, activeUsers)
			log.Error(err)
		}
	}
}

func handleNetworkPoll(ctx context.Context, id int, c cache.Cache, contract types.Contract, notify BalanceNotifierFunc, lastCheckedBlock uint64, latestBlock uint64, client Client, activeUsers []common.Address) (uint64, error) {
	var err error
	var fromBlock uint64

	if lastCheckedBlock == 0 {
		// the first time this loop is running, only check latest block
		fromBlock = latestBlock
	} else {
		// otherwise, add 1 to last checked block to not include it
		fromBlock = lastCheckedBlock + 1
	}
	toBlock := fromBlock // check one block at a time

	err = processLogs(ctx, id, c, contract, notify, fromBlock, toBlock, activeUsers)
	if err != nil {
		return lastCheckedBlock, err
	}

	return toBlock, nil
}

func processLogs(ctx context.Context, id int, c cache.Cache, contract types.Contract, notify BalanceNotifierFunc, fromBlock uint64, toBlock uint64, userAddresses []common.Address) error {
	balances := make(map[string]*cache.UserBalance)

	logs, err := contract.GetLogs(
		&bind.FilterOpts{
			Start: fromBlock,
			End:   &toBlock,
		}, userAddresses,
	)

	if len(logs) == 0 {
		return nil
	}

	originalBalances, err := c.GetRuleTokenBalance(ctx, id)
	if err != nil {
		return err
	}

	log.Debug("rule:", id, "-", len(logs), "logs found from block:", fromBlock, "to:", toBlock)
	if err != nil {
		return err
	}
	for _, logItem := range logs {
		old := originalBalances[logItem.User.String()]
		if old == nil {
			old = &cache.UserBalance{
				Balance:     big.NewInt(0),
				BlockNumber: 0,
			}
		}
		// if the cache contains a newer balance, don't update it
		if old.BlockNumber == 0 || toBlock > old.BlockNumber {
			balances[logItem.User.String()] = &cache.UserBalance{
				Balance:     old.Balance.Add(old.Balance, logItem.Value),
				BlockNumber: toBlock,
			}
		}
	}
	updatedBalances, err := c.UpdateRuleTokenBalances(ctx, &cache.TokenBalances{
		RuleID:      id,
		Balances:    balances,
		BlockNumber: toBlock,
	})
	if err != nil {
		return err
	}
	for user, balance := range updatedBalances.Balances {
		notify(id, user, balance.Balance)
	}
	return nil
}
