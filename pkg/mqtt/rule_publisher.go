package mqtt

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/momentum-xyz/token-monitor/pkg/log"
	"github.com/pkg/errors"
)

type RulePublisher struct {
	client mqtt.Client
}

func NewRuleStatePublisher(client mqtt.Client) *RulePublisher {
	return &RulePublisher{
		client: client,
	}
}

func (p *RulePublisher) PublishRuleState(id int, user string, satisfied bool) {
	err := Publish(p.client, "permissions", 1, fmt.Sprintf(`"accountAddress": %s, "%d": %v, `, user, id, satisfied))
	// TODO retry publishes if they fail?
	log.Error(errors.WithMessagef(err, "failed to publish rule state %d for user %s", id, user))
}
