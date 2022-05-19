package networks

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/OdysseyMomentumExperience/token-service/pkg/constants"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ory/x/errorsx"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"
)

type ClientManager struct {
	transport http.RoundTripper
	ethereum  map[string]*EthereumLoadBalancer

	networkTypes map[string]string
}

func NewClientManager(cfg *Config) (*ClientManager, error) {
	networkTypes := make(map[string]string)
	for _, network := range cfg.Networks {
		networkTypes[network.Name] = network.Type
	}

	cm := &ClientManager{
		ethereum:     make(map[string]*EthereumLoadBalancer),
		transport:    newTransport(),
		networkTypes: networkTypes,
	}
	var err error

	for _, network := range cfg.Networks {
		switch network.Type {
		case constants.NetworkTypeEthereum:
			err = cm.RegisterEthereumNetwork(network)

		default:
			return nil, errors.Errorf("unknown network type %q", network.Type)
		}

		if err != nil {
			return nil, err
		}
	}

	return cm, err
}

func (c *ClientManager) Close() {
	for _, clients := range c.ethereum {
		for _, client := range clients.clients {
			client.Close()
		}
	}
}

func (c *ClientManager) RegisterEthereumNetwork(networkConfig *NetworkConfig) error {
	lb, err := NewEthereumLoadBalancer(networkConfig, c.transport)
	if err != nil {
		return err

	}
	c.ethereum[networkConfig.Name] = lb
	return nil
}

func (c *ClientManager) GetNetworkType(networkName string) string {
	return c.networkTypes[networkName]
}

func (c *ClientManager) GetEthereumClient(networkName string) (*ethclient.Client, error) {
	clients, ok := c.ethereum[networkName]

	if !ok {
		return nil, errors.Errorf("no client is registered for network %q", networkName)
	}

	return clients.GetClient(), nil
}

func newTransport() *http.Transport {
	return &http.Transport{
		// MaxIdleConnsPerHost: 100,
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxConnsPerHost:       2,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
}

func newThrottledHTTPClient(cfg *HostConfig, transport http.RoundTripper) *http.Client {
	// TODO use an http client that supports retries?
	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: newThrottledTransport(cfg, transport),
	}
}

func newThrottledTransport(cfg *HostConfig, transport http.RoundTripper) http.RoundTripper {
	var lim *rate.Limiter
	if cfg.RateLimit != 0 {
		if cfg.BurstLimit == 0 {
			cfg.BurstLimit = 50
		}
		lim = rate.NewLimiter(cfg.RateLimit, cfg.BurstLimit)
	}

	metricsNamespace := "aa_http"

	u, _ := url.Parse(cfg.URL)
	safeURL := u.Redacted()

	nOK := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "successful_requests",
		ConstLabels: prometheus.Labels{
			"url": safeURL,
		},
	})

	okLatencies := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Name:      "success_response_time",
		Help:      "Success response time",
		ConstLabels: prometheus.Labels{
			"url": safeURL,
		},
	})

	errLatencies := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Name:      "error_response_time",
		Help:      "Error response time",
		ConstLabels: prometheus.Labels{
			"url": safeURL,
		},
	})

	waitLatencies := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Name:      "waiting_time",
		Help:      "Error response time",
		ConstLabels: prometheus.Labels{
			"url": safeURL,
		},
	})
	nErr := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		// Subsystem: cfg.URL,
		Name: "failed_requests",
		ConstLabels: prometheus.Labels{
			"url": safeURL,
		},
	})

	return &ThrottledTransport{
		transport:     transport,
		lim:           lim,
		nOK:           nOK,
		nErr:          nErr,
		okLatencies:   okLatencies,
		errLatencies:  errLatencies,
		waitLatencies: waitLatencies,
	}
}

type ThrottledTransport struct {
	transport http.RoundTripper
	lim       *rate.Limiter

	nOK           prometheus.Counter
	nErr          prometheus.Counter
	okLatencies   prometheus.Histogram
	errLatencies  prometheus.Histogram
	waitLatencies prometheus.Histogram
}

func (t *ThrottledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tt := time.Now()
	if err := t.lim.Wait(req.Context()); err != nil {
		return nil, err
	}
	t.waitLatencies.Observe(time.Since(tt).Seconds())

	res, err := t.transport.RoundTrip(req)

	if err != nil {
		t.nErr.Inc()
		t.errLatencies.Observe(time.Since(tt).Seconds())
		return nil, err
	}

	t.nOK.Inc()
	t.okLatencies.Observe(time.Since(tt).Seconds())
	return res, nil
}

func newRPCClient(url string, hc *http.Client) (*rpc.Client, error) {
	rc, err := rpc.DialHTTPWithClient(url, hc)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}
	return rc, nil
}

func newGethClient(url string, hc *http.Client) (*ethclient.Client, error) {
	rc, err := newRPCClient(url, hc)
	if err != nil {
		return nil, err
	}

	return ethclient.NewClient(rc), nil
}
