package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/OdysseyMomentumExperience/token-service/pkg/cache"
	"github.com/go-redis/redis/v8"
	"github.com/ory/x/errorsx"
	"github.com/vmihailenco/msgpack/v5"
	"math/big"
)

type Cache struct {
	client *redis.Client
}

func NewCache(cfg *cache.Config) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:        cfg.Redis.Addr,
		Network:     cfg.Redis.Network,
		DB:          cfg.Redis.DB,
		DialTimeout: cfg.Redis.ConnectTimeout,
		ReadTimeout: cfg.Redis.ReadTimeout,
		Username:    cfg.Redis.Username,
		Password:    cfg.Redis.Password,
		IdleTimeout: cfg.Redis.IdleTimeout,
	})

	return &Cache{
		client: client,
	}
}

func (c *Cache) ClearRuleTokenBalances(ctx context.Context, ruleID int) error {
	key := cacheKey(ruleID)
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) SetRuleTokenBalances(ctx context.Context, tb *cache.TokenBalances) error {
	key := cacheKey(tb.RuleID)
	return c.setRuleTokenBalances(ctx, key, tb)
}

func (c *Cache) setRuleTokenBalances(ctx context.Context, key string, tb *cache.TokenBalances) error {
	totals, err := c.getRuleTokenBalance(ctx, key)
	if err != nil {
		return err
	}

	for user, balance := range tb.Balances {
		totals[user] = balance
	}

	err = c.set(ctx, key, totals)
	if err != nil {
		return errorsx.WithStack(err)
	}

	return nil
}

func (c *Cache) UpdateRuleTokenBalances(ctx context.Context, tb *cache.TokenBalances) (*cache.TokenBalances, error) {
	key := cacheKey(tb.RuleID)
	var res *cache.TokenBalances

	txf := func(tx *redis.Tx) error {
		var err error
		res, err = c.updateRuleTokenBalances(ctx, key, tb)
		return err
	}

	for i := 0; i < 10; i++ {
		err := c.client.Watch(ctx, txf, key)
		if err == nil {
			return res, nil
		}
		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}
		return nil, errorsx.WithStack(err)
	}

	return nil, errors.New("update reached maximum number of retries")
}

func (c *Cache) updateRuleTokenBalances(ctx context.Context, key string, tb *cache.TokenBalances) (*cache.TokenBalances, error) {
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
func (c *Cache) getRuleTokenBalance(ctx context.Context, key string) (map[string]*cache.UserBalance, error) {
	tb := &cache.TokenBalances{
		Balances: make(map[string]*cache.UserBalance),
	}
	err := c.get(ctx, key, tb)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	return tb.Balances, nil
}

func cacheKey(id int) string {
	return fmt.Sprintf("tokensvc:rule:%v:balances", id)
}

func (c *Cache) Close() {
	c.client.Close()
}

func (c *Cache) set(ctx context.Context, key string, v interface{}) error {
	b, err := msgpack.Marshal(v)
	if err != nil {
		return errorsx.WithStack(err)
	}

	return errorsx.WithStack(c.client.Set(ctx, key, b, 0).Err())
}

func (c *Cache) get(ctx context.Context, key string, v interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil && err != redis.Nil {
		return errorsx.WithStack(err)
	}

	err = msgpack.Unmarshal([]byte(data), v)
	if err != nil {
		return errorsx.WithStack(err)
	}

	return nil
}
