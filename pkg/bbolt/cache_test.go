package bbolt

import (
	"context"
	"github.com/OdysseyMomentumExperience/token-service/pkg/cache"
	"math/big"
	"math/rand"
	"testing"

	"github.com/instana/testify/require"
)

func Test_NewCache(t *testing.T) {
	_, err := NewCache(&cache.Config{Type: "bbolt", BBolt: cache.BBoltConfig{
		BucketName: "test", Path: "test.db", FileMode: 0600,
	}})
	require.NoError(t, err, "failed to create cache")
}

func TestCache_SetAndGetRuleTokenBalances(t *testing.T) {
	c, err := NewCache(&cache.Config{Type: "bbolt", BBolt: cache.BBoltConfig{
		BucketName: "test", Path: "test.db", FileMode: 0600,
	}})
	require.NoError(t, err, "could not create cache")

	ctx := context.Background()
	tb := &cache.TokenBalances{
		RuleID:      rand.Int(),
		BlockNumber: 100000,
		Balances: map[string]*cache.UserBalance{
			"0x1": {Balance: big.NewInt(1)},
			"0x2": {Balance: big.NewInt(2)},
			"0x3": {Balance: big.NewInt(3)},
			"0x4": {Balance: big.NewInt(4)},
		},
	}

	err = c.SetRuleTokenBalances(ctx, tb)
	require.NoError(t, err, "could not set token balances")

	got, err := c.GetRuleTokenBalance(ctx, tb.RuleID)
	require.NoError(t, err, "could not get token balances")

	require.Equal(t, tb.Balances, got)
}
