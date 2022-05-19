package mqtt

import (
	"strconv"
	"time"

	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/pkg/errors"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const mqttClientId = "token-service"

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Logln(0, "Connected to MQTT broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Logf(0, "Connection to MQTT broker lost: %v\n", err)
	client.Connect()
}

func mqttConnect(cfg *Config) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://" + cfg.HOST + ":" + strconv.FormatUint(uint64(cfg.PORT), 10))
	opts.SetClientID("env-control" + mqttClientId)
	opts.SetUsername(cfg.USER)
	opts.SetPassword(cfg.PASSWORD)
	// opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(2 * time.Second)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}

func GetMQTTClient(cfg *Config) (mqtt.Client, error) {
	return mqttConnect(cfg)
}

func Subscribe(c mqtt.Client, topic string, qos int, channel chan []byte) error {
	if token := c.Subscribe(topic, byte(qos), func(client mqtt.Client, msg mqtt.Message) {
		channel <- msg.Payload()
	}); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func Publish(c mqtt.Client, topic string, qos int, msg string) error {
	if token := c.Publish(topic, byte(qos), false, msg); token.Wait() && token.Error() != nil {
		return errors.Wrap(token.Error(), "failed to publish message")
	}
	return nil
}
