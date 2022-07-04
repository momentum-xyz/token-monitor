package bbolt

import (
	"context"
	"errors"
	"fmt"
	"github.com/momentum-xyz/token-monitor/pkg/cache"
	"github.com/ory/x/errorsx"
	"github.com/vmihailenco/msgpack/v5"
	bolt "go.etcd.io/bbolt"
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

	err := c.set(ctx, key, tb)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	return tb, nil
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
		b := tx.Bucket([]byte(c.config.BBolt.BucketName))
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
	err := c.client.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(c.config.BBolt.BucketName))
		if err != nil {
			return errors.New(fmt.Sprintf("Error fetching bucket: %s", err))
		}
		data = b.Get(key)

		return nil
	})

	if err != nil {
		return err
	}
	if data == nil || len(data) == 0 {
		return nil
	}
	err = msgpack.Unmarshal(data, v)
	if err != nil {
		return errorsx.WithStack(err)
	}

	return nil
}
