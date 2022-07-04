package redis

import (
	"context"
	"github.com/momentum-xyz/token-monitor/pkg/cache"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/instana/testify/require"
)

func TestCache_UpdateRuleTokenBalances(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	type fields struct {
		Cache *Cache
	}
	type args struct {
		ctx context.Context
		tb  *cache.TokenBalances
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				Cache: NewCache(
					&cache.Config{
						Type:  "redis",
						Redis: cache.RedisConfig{Addr: "localhost:6379"},
					},
				),
			},
			args: args{
				ctx: context.Background(),
				tb: &cache.TokenBalances{
					RuleID:      rand.Int(),
					BlockNumber: 100000,
					Balances: map[string]*cache.UserBalance{
						"0x1": {Balance: big.NewInt(1)},
						"0x2": {Balance: big.NewInt(2)},
						"0x3": {Balance: big.NewInt(3)},
						"0x4": {Balance: big.NewInt(4)},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 4; i++ {
				got, err := tt.fields.Cache.UpdateRuleTokenBalances(tt.args.ctx, tt.args.tb)
				if (err != nil) != tt.wantErr {
					t.Errorf("Cache.UpdateRuleTokenBalances() error = %+v, wantErr %+v", err, tt.wantErr)
					return
				}

				balances := make(map[string]*cache.UserBalance)
				for k, v := range tt.args.tb.Balances {
					bb := new(big.Int).Mul(v.Balance, big.NewInt(int64(i+1)))
					balances[k] = &cache.UserBalance{Balance: bb}
				}
				want := &cache.TokenBalances{
					RuleID:      tt.args.tb.RuleID,
					BlockNumber: tt.args.tb.BlockNumber,
					Balances:    balances,
				}

				if !reflect.DeepEqual(got, want) {
					t.Errorf("Cache.UpdateRuleTokenBalances() = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestCache_UpdateRuleTokenBalances2(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	id := rand.Int()

	type fields struct {
		Cache *Cache
	}
	type args struct {
		ctx context.Context
		tb  []*cache.TokenBalances
	}
	type want struct {
		tb []*cache.TokenBalances
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				Cache: NewCache(
					&cache.Config{
						Type:  "redis",
						Redis: cache.RedisConfig{Addr: "localhost:6379"},
					},
				),
			},
			args: args{
				ctx: context.Background(),
				tb: []*cache.TokenBalances{
					{RuleID: id, BlockNumber: 100000, Balances: map[string]*cache.UserBalance{"0x1": {Balance: big.NewInt(1)}, "0x2": {Balance: big.NewInt(2)}, "0x3": {Balance: big.NewInt(3)}, "0x4": {Balance: big.NewInt(4)}}},
					{RuleID: id, BlockNumber: 100001, Balances: map[string]*cache.UserBalance{"0x2": {Balance: big.NewInt(2)}}},
					{RuleID: id, BlockNumber: 100002, Balances: map[string]*cache.UserBalance{"0x3": {Balance: big.NewInt(3)}, "0x4": {Balance: big.NewInt(4)}}},
					{RuleID: id, BlockNumber: 100003, Balances: map[string]*cache.UserBalance{"0x1": {Balance: big.NewInt(3)}}},
					{RuleID: id, BlockNumber: 100003, Balances: map[string]*cache.UserBalance{"0x2": {Balance: big.NewInt(3)}}},
				},
			},
			want: want{
				tb: []*cache.TokenBalances{
					{RuleID: id, BlockNumber: 100000, Balances: map[string]*cache.UserBalance{"0x1": {Balance: big.NewInt(1)}, "0x2": {Balance: big.NewInt(2)}, "0x3": {Balance: big.NewInt(3)}, "0x4": {Balance: big.NewInt(4)}}},
					{RuleID: id, BlockNumber: 100001, Balances: map[string]*cache.UserBalance{"0x2": {Balance: big.NewInt(4)}}},
					{RuleID: id, BlockNumber: 100002, Balances: map[string]*cache.UserBalance{"0x3": {Balance: big.NewInt(6)}, "0x4": {Balance: big.NewInt(8)}}},
					{RuleID: id, BlockNumber: 100003, Balances: map[string]*cache.UserBalance{"0x1": {Balance: big.NewInt(4)}}},
					{RuleID: id, BlockNumber: 100003, Balances: map[string]*cache.UserBalance{"0x2": {Balance: big.NewInt(7)}}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < len(tt.args.tb); i++ {
				got, err := tt.fields.Cache.UpdateRuleTokenBalances(tt.args.ctx, tt.args.tb[i])
				if (err != nil) != tt.wantErr {
					t.Errorf("Cache.UpdateRuleTokenBalances() error = %+v, wantErr %+v", err, tt.wantErr)
					return
				}

				require.Equal(t, tt.want.tb[i], got)
				if !reflect.DeepEqual(got, tt.want.tb[i]) {
					t.Errorf("Cache.UpdateRuleTokenBalances() = %v, want %v", got, tt.want.tb[i])
				}
			}
		})
	}
}
