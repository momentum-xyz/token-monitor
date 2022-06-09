package mqtt

import (
	"context"

	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/OdysseyMomentumExperience/token-service/pkg/web3"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ory/x/errorsx"
)

type RuleBroker struct {
	client mqtt.Client

	registrationChannel chan []byte
	rulesChannel        chan []byte
	activeRulesChannel  chan []byte
	activeUsersChannel  chan []byte

	ruleManager *web3.RuleManager
}

func NewRuleBroker(client mqtt.Client, ruleManager *web3.RuleManager, cfg *Config) (*RuleBroker, error) {
	registrationChannel := make(chan []byte)
	rulesChannel := make(chan []byte)
	activeRulesChannel := make(chan []byte)
	activeUsersChannel := make(chan []byte)

	mqttTopics := cfg.TOPICS

	err := Subscribe(client, mqttTopics.ActiveUsersTopic, 2, activeUsersChannel)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	err = Subscribe(client, mqttTopics.ActiveRulesTopic, 2, activeRulesChannel)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	err = Subscribe(client, mqttTopics.RegistrationTopic, 2, registrationChannel)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	err = Subscribe(client, mqttTopics.RulesTopic, 2, rulesChannel)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}

	return &RuleBroker{
		client:              client,
		registrationChannel: registrationChannel,
		rulesChannel:        rulesChannel,
		activeRulesChannel:  activeRulesChannel,
		activeUsersChannel:  activeUsersChannel,
		ruleManager:         ruleManager,
	}, nil
}

func (p *RuleBroker) Start(ctx context.Context) {
	go p.start(ctx)
}

func (p *RuleBroker) start(ctx context.Context) {
	log.Logln(0, "rule broker -", "starting rule broker...")

	err := p.initializeRuleManager(ctx)
	log.Error(err)
	log.Logln(0, "rule broker -", "starting rule broker... done")

	for {
		var err error
		select {
		case <-ctx.Done():
			log.Error(err)
			return
		case userJSON := <-p.registrationChannel:
			err = handleNewUser(ctx, userJSON, p.ruleManager)
		case ruleJSON := <-p.rulesChannel:
			err = handleNewRule(ctx, ruleJSON, p.ruleManager)
		}

		log.Error(err)
	}
}

func (p *RuleBroker) initializeRuleManager(ctx context.Context) error {
	log.Logln(0, "rule broker -", "waiting for initial rules and users")
	
	//TODO: refactor structure of message for users coming from backend service - for discussion
	users, err := web3.UnmarshalUsers(<-p.activeUsersChannel)
	if err != nil {
		return errorsx.WithStack(err)
	}
	log.Logln(0, "rule broker -", "users list created")

	rules, err := web3.UnmarshalRuleDefinitions(<-p.activeRulesChannel)
	if err != nil {
		return errorsx.WithStack(err)
	}
	log.Logln(0, "rule broker -", "rules list created")

	return p.ruleManager.Init(ctx, rules, users)
}

func handleNewRule(ctx context.Context, ruleJSON []byte, manager *web3.RuleManager) error {
	rule, err := web3.UnmarshalRuleDefinition(ruleJSON)
	if err != nil {
		return errorsx.WithStack(err)
	}
	log.Logln(0, "rule broker -", "received rule:", rule)

	if !rule.Active {
		log.Logln(0, "rule broker -", "deleting rule with id:", rule.ID)

		manager.DeleteRule(ctx, rule.ID)
		return nil
	}
	err = manager.StartNewTokenTracker(ctx, rule)
	if err != nil {
		return errorsx.WithStack(err)
	}

	return nil
}

func handleNewUser(ctx context.Context, userJSON []byte, ruleManager *web3.RuleManager) error {
	user, err := web3.UnmarshalUser(userJSON)
	if err != nil {
		return errorsx.WithStack(err)
	}
	log.Logln(0, "rule broker -", "new address received", user.EthereumAddress)

	return ruleManager.AddUser(user)
}
