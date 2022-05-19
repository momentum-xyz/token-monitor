package main

import (
	"os"

	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/OdysseyMomentumExperience/token-service/pkg/mqtt"
	"github.com/OdysseyMomentumExperience/token-service/pkg/tokensvc"
)

const configFileName = "config.dev.yaml"

func main() {
	cfg := tokensvc.LoadConfig(configFileName)
	cfg.PrettyPrint()

	// token-service is a identifier for this client
	mqttClient, err := mqtt.GetMQTTClient(cfg.MQTT)

	if err != nil {
		log.Logln(0, err)
		os.Exit(0)
	}
	defer mqttClient.Disconnect(5)

	commsFromOdyssey := make(chan []byte)

	// token-service will listen for messages coming for this topic
	subscriptionTopic := "$share/token-service"
	err = mqtt.Subscribe(mqttClient, subscriptionTopic, 1, commsFromOdyssey)

	if err != nil {
		log.Logln(0, err)
	}

	for {
		select {
		case msg := <-commsFromOdyssey:
			log.Logln(0, msg)
			continue
		}
	}
}
