package bbolt

import (
	"context"
	"errors"
	"fmt"
	"github.com/momentum-xyz/token-monitor/pkg/cache"
	"github.com/ory/x/errorsx"
	"github.com/vmihailenco/msgpack/v5"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"os"
)

type Cache struct {
	client *bolt.DB
	config *cache.Config
}

func NewCache(cfg *cache.Config) (*Cache, error) {
	client, err := bolt.Open(cfg.BBolt.Path, os.FileMode(cfg.BBolt.FileMode), nil)
	if err != nil {
		return nil, err
	}
	return &Cache{
		client: client,
		config: cfg,
	}, nil
}

func (c *Cache) ClearRuleTokenBalances(ctx context.Context, ruleID int) error {
	key := cacheKey(ruleID)
	return c.client.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(c.config.BBolt.BucketName))
		if err != nil {
			return err
		}
		err = b.Delete(key)
		if err != nil {
			return err
		}
		return nil
	})
}

func (c *Cache) SetRuleTokenBalances(ctx context.Context, tb *cache.TokenBalances) error {
	key := cacheKey(tb.RuleID)
	return c.setRuleTokenBalances(ctx, key, tb)
}

func (c *Cache) setRuleTokenBalances(ctx context.Context, key []byte, tb *cache.TokenBalances) error {
	// FIXME consider to use an update here to not open tx twice
	totals, err := c.getRuleTokenBalance(ctx, key)
	if err != nil {
		return err
	}

	for user, balance := range tb.Balances {
		totals[user] = balance
	}

	err = c.set(ctx, key, &cache.TokenBalances{
		RuleID:      tb.RuleID,
		BlockNumber: tb.BlockNumber,
		Balances:    totals,
	})
	if err != nil {
		return errorsx.WithStack(err)
	}

	return nil
}

func (c *Cache) UpdateRuleTokenBalances(ctx context.Context, tb *cache.TokenBalances) (*cache.TokenBalances, error) {
	key := cacheKey(tb.RuleID)
	res, err := c.updateRuleTokenBalances(ctx, key, tb)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Cache) updateRuleTokenBalances(ctx context.Context, key []byte, tb *cache.TokenBalances) (*cache.TokenBalances, error) {
	totals, err := c.getRuleTokenBalance(ctx, key)
	if err != nil {
		return nil, err
	}

	res := &cache.TokenBalances{
		RuleID:      tb.RuleID,
		BlockNumber: tb.BlockNumber,
		Balances:    make(map[string]*cache.UserBalance),
	}

	for user, newUserBalance := range tb.Balances {
		oldUserBalance := totals[user]
		if oldUserBalance == nil {
			oldUserBalance = &cache.UserBalance{
				Balance:     big.NewInt(0),
				BlockNumber: 0,
			}
		}
		// if the cache contains a newer balance, don't update it
		if newUserBalance.BlockNumber == 0 || newUserBalance.BlockNumber > oldUserBalance.BlockNumber {
			newBalance := new(big.Int).Add(oldUserBalance.Balance, newUserBalance.Balance)
			totals[user] = &cache.UserBalance{
				Balance:     new(big.Int).Set(newBalance),
				BlockNumber: newUserBalance.BlockNumber,
			}
			res.Balances[user] = &cache.UserBalance{
				Balance:     new(big.Int).Set(newBalance),
				BlockNumber: newUserBalance.BlockNumber,
			}
		}
	}

	totalsTb := &cache.TokenBalances{
		Balances:    totals,
		BlockNumber: tb.BlockNumber,
		RuleID:      tb.RuleID,
	}
	res.BlockNumber = tb.BlockNumber
	err = c.set(ctx, key, totalsTb)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	return res, nil
}

func (c *Cache) GetRuleTokenBalance(ctx context.Context, id int) (map[string]*cache.UserBalance, error) {
	return c.getRuleTokenBalance(ctx, cacheKey(id))
}
func (c *Cache) getRuleTokenBalance(ctx context.Context, key []byte) (map[string]*cache.UserBalance, error) {
	tb := &cache.TokenBalances{
		Balances: make(map[string]*cache.UserBalance),
	}
	err := c.get(ctx, key, tb)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	return tb.Balances, nil
}

func cacheKey(id int) []byte {
	return []byte(fmt.Sprintf("tokensvc:rule:%v:balances", id))
}

func (c *Cache) Close() {
	c.client.Close()
}

func (c *Cache) set(ctx context.Context, key []byte, v interface{}) error {
	val, err := msgpack.Marshal(v)
	if err != nil {
		return errorsx.WithStack(err)
	}
	c.client.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(c.config.BBolt.BucketName))
		if err != nil {
			return err
		}
		err = b.Put(key, val)
		if err != nil {
			return err
		}
		return nil
	})

	return nil
}

func (c *Cache) get(ctx context.Context, key []byte, v interface{}) error {
	var data []byte
	err := c.client.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.config.BBolt.BucketName))
		if b == nil {
			return errors.New(fmt.Sprintf("No result for bucket: %s", string(c.config.BBolt.BucketName)))
		}
		data = b.Get(key)
		if data == nil {
			return errors.New(fmt.Sprintf("No result for key: %s", string(key)))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	err = msgpack.Unmarshal(data, v)
	if err != nil {
		return errorsx.WithStack(err)
	}

	return nil
}
