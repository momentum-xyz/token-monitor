package cache

import (
	"context"
	"math/big"
	"time"
)

type Cache interface {
	GetRuleTokenBalance(ctx context.Context, id int) (map[string]*UserBalance, error)
	SetRuleTokenBalances(ctx context.Context, tb *TokenBalances) error
	ClearRuleTokenBalances(ctx context.Context, ruleID int) error
	UpdateRuleTokenBalances(ctx context.Context, tb *TokenBalances) (*TokenBalances, error)
	Close()
}

type UserBalance struct {
	Balance     *big.Int
	BlockNumber uint64
}

type TokenBalances struct {
	RuleID      int
	BlockNumber uint64
	Balances    map[string]*UserBalance
}

type Config struct {
	Type  string      `yaml:"type"`
	BBolt BBoltConfig `yaml:"bbolt"`
	Redis RedisConfig `yaml:"redis"`
}
type BBoltConfig struct {
	Path       string `yaml:"path"`
	FileMode   uint32 `yaml:"fileMode"`
	BucketName string `yaml:"bucketName"`
}

type RedisConfig struct {
	Network        string        `json:"network"`
	Addr           string        `json:"addr"`
	Username       string        `yaml:"username"`
	Password       string        `yaml:"password"`
	DB             int           `yaml:"db"`
	ConnectTimeout time.Duration `yaml:"connect_timeout"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
}
