package eth

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/momentum-xyz/token-monitor/pkg/log"
	"github.com/pkg/errors"
)

func startBlockChecker(ctx context.Context, id int, client *ethclient.Client, blockCh ...chan<- uint64) {
	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	var latestBlock uint64 = 0

	for {
		var err error

		select {
		case <-ctx.Done():
			log.Logln(0, "rule:", id, "-", "Stopping active user manager", ctx.Err())
			return
		case <-ticker.C:
			latestBlock, err = handleBlockPoll(ctx, id, client, latestBlock, blockCh...)
		}

		log.Error(err)
	}
}

func handleBlockPoll(ctx context.Context, id int, client *ethclient.Client, latestBlock uint64, blockCh ...chan<- uint64) (uint64, error) {
	var nextBlock uint64

	nextBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return latestBlock, errors.Wrapf(err, "rule: %d - Error fetching latest block number", id)
	}
	if nextBlock == latestBlock {
		log.Logln(0, "rule:", id, "-", "No new block")
		return latestBlock, nil
	}

	for _, ch := range blockCh {
		select {
		case ch <- nextBlock:
		default:
		}
	}

	return nextBlock, nil

}
