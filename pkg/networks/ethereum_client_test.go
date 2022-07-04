package networks_test

import (
	"fmt"
	"math"
	"math/big"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/instana/testify/require"
	"github.com/momentum-xyz/token-monitor/pkg/abigen"
	"github.com/momentum-xyz/token-monitor/pkg/networks"
)

const infuraURL = "https://ropsten.infura.io/v3/a6bd0a4b91074054a7d17e27d85d9fac"
const contractAddr = "0xc778417E063141139Fce010982780140Aa0cD5Ab"
const userAddr = "0xa428d6424f9430b293eccc49fa71971b8a9d258d"

func TestEthereumClient(t *testing.T) {
	tt := time.Now()
	c, err := networks.NewThrottledEthereumClient(&networks.HostConfig{
		URL:        infuraURL,
		RateLimit:  2,
		BurstLimit: 1,
	}, http.DefaultTransport)
	require.NoError(t, err)

	wg := &sync.WaitGroup{}

	token, err := abigen.NewERC20Caller(common.HexToAddress(contractAddr), c)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := erc20Balance(token, nil, userAddr)
			require.NoError(t, err)

			fmt.Println(res, time.Now())
		}()
	}

	wg.Wait()

	tt2 := time.Now()
	dur := tt2.Sub(tt)
	fmt.Println(dur)

	require.WithinDuration(t, tt.Add(5*time.Second), tt2, time.Second)
}

func erc20Balance(token *abigen.ERC20Caller, blockNumber *big.Int, userAddr string) (string, error) {
	amt, err := token.BalanceOf(&bind.CallOpts{
		BlockNumber: blockNumber,
	}, common.HexToAddress(userAddr))
	if err != nil {
		return "", err
	}

	fBalance := new(big.Float)
	fBalance.SetString(amt.String())
	bal := new(big.Float).Quo(fBalance, big.NewFloat(math.Pow10(18)))

	return bal.String(), nil
}
