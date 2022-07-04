package tokensvc

import (
	"context"
	"github.com/momentum-xyz/token-monitor/pkg/bbolt"
	"github.com/momentum-xyz/token-monitor/pkg/cache"
	"github.com/momentum-xyz/token-monitor/pkg/redis"
	"github.com/momentum-xyz/token-monitor/pkg/server"
	"github.com/momentum-xyz/token-monitor/pkg/web3/eth"
	"github.com/pkg/errors"

	"github.com/momentum-xyz/token-monitor/pkg/log"
	"github.com/momentum-xyz/token-monitor/pkg/mqtt"
	"github.com/momentum-xyz/token-monitor/pkg/networks"
	"github.com/momentum-xyz/token-monitor/pkg/web3"
)

type RuleBroker interface {
	Start(ctx context.Context)
}

type MetricsServer interface {
	Start(ctx context.Context)
}

type APIServer interface {
	Start(ctx context.Context)
}

type TokenService struct {
	RuleBroker    RuleBroker
	MetricsServer MetricsServer
	APIServer     APIServer
}

func NewTokenService(ruleBroker RuleBroker, metricsServer MetricsServer, apiServer APIServer) *TokenService {
	return &TokenService{
		RuleBroker:    ruleBroker,
		MetricsServer: metricsServer,
		APIServer:     apiServer,
	}
}

func NewMQTTTokenService(cfg *Config) (*TokenService, func(), error) {
	cleanup := newCleanup()

	log.SetConfig(cfg.Log)

	var c cache.Cache
	var err error
	switch cfg.Cache.Type {
	case "bbolt":
		c, err = bbolt.NewCache(cfg.Cache)
		if err != nil {
			cleanup.do()
			return nil, nil, err
		}
	case "redis":
		c = redis.NewCache(cfg.Cache)
	default:
		return nil, nil, errors.New("could not find valid caching option in configuration")
	}
	cleanup.add(c.Close)

	mqttClient, err := mqtt.GetMQTTClient(cfg.MQTT)
	if err != nil {
		cleanup.do()
		return nil, nil, err
	}
	cleanup.add(func() { mqttClient.Disconnect(5) })
	networkManager, err := networks.NewClientManager(cfg.NetworkManager)
	if err != nil {
		cleanup.do()
		return nil, nil, err
	}
	cleanup.add(networkManager.Close)

	ruleStatePublisher := mqtt.NewRuleStatePublisher(mqttClient)
	ruleManager := web3.NewRuleManager(networkManager, ruleStatePublisher, c)
	mqttServer, err := mqtt.NewRuleBroker(mqttClient, ruleManager, cfg.MQTT)
	if err != nil {
		cleanup.do()
		return nil, nil, err
	}

	metricsServer := NewPrometheusServer()

	tokenNameService := eth.NewEthNameService(networkManager)
	tokenNameHandler := server.NewTokenNameHandler(tokenNameService)
	apiServer := server.NewServer(tokenNameHandler)

	return NewTokenService(mqttServer, metricsServer, apiServer), cleanup.do, nil
}

func (svc *TokenService) Start(ctx context.Context) error {
	svc.MetricsServer.Start(ctx)
	svc.RuleBroker.Start(ctx)
	svc.APIServer.Start(ctx)

	return nil
}

func newCleanup() *cleanup {
	return &cleanup{make([]func(), 0)}
}

type cleanup struct {
	fns []func()
}

func (c *cleanup) add(fn func()) {
	c.fns = append(c.fns, fn)
}
func (c *cleanup) do() {
	for i := len(c.fns) - 1; i >= 0; i-- {
		c.fns[i]()
	}
}
