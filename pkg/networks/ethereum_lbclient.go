package networks

import (
	"math/rand"
	"net/http"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ory/x/errorsx"
)

type EthereumLoadBalancer struct {
	current uint64
	clients []*ethclient.Client
}

func NewThrottledEthereumClient(hostConfig *HostConfig, transport http.RoundTripper) (*ethclient.Client, error) {
	client, err := newGethClient(hostConfig.URL, newThrottledHTTPClient(hostConfig, transport))
	if err != nil {
		return nil, errorsx.WithStack(err)
	}
	return client, nil
}

func NewEthereumLoadBalancer(networkConfig *NetworkConfig, transport http.RoundTripper) (*EthereumLoadBalancer, error) {
	clients := make([]*ethclient.Client, len(networkConfig.Hosts))
	var err error

	for i, hostConfig := range networkConfig.Hosts {
		clients[i], err = newGethClient(hostConfig.URL, newThrottledHTTPClient(hostConfig, transport))

		if err != nil {
			return nil, errorsx.WithStack(err)
		}
	}
	return &EthereumLoadBalancer{
		clients: clients,
	}, nil
}

func (emc *EthereumLoadBalancer) GetClient() *ethclient.Client {
	// TODO: better load balancing? currently we use round robin.
	return emc.getNextClient()
}
func (emc *EthereumLoadBalancer) getNextClient() *ethclient.Client {
	return emc.clients[emc.nextIndex()]
}

func (emc *EthereumLoadBalancer) nextIndex() int {
	return int(atomic.AddUint64(&emc.current, uint64(1)) % uint64(len(emc.clients)))
}

func (emc *EthereumLoadBalancer) getRandomClient() *ethclient.Client {
	return emc.clients[rand.Intn(len(emc.clients))]
}
