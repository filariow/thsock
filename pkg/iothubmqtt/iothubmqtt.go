package iothubmqtt

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient interface {
	Publish(topic, message string) error
}

func NewMQTTClient(config *Config) MQTTClient {
	return &mqttClient{config: *config}
}

type mqttClient struct {
	config Config
}

func (c *mqttClient) Publish(topic, message string) error {
	cfg := c.config
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", cfg.Broker, cfg.Port))
	opts.SetClientID(cfg.ClientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)

	opts.OnConnect = func(client mqtt.Client) { log.Printf("MQTT Client to '%s' connected", cfg.Broker) }
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("MQTT Client lost connection to '%s': '%s'", cfg.Broker, err)
	}

	client := mqtt.NewClient(opts)
	// throw an error if the connection isn't successfull
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Failed to connect: %w", token.Error())
	}

	token := client.Publish(topic, 0, false, message)
	<-token.Done()
	if token.Error() != nil {
		return fmt.Errorf("Failed to publish to topic '%s': %w", topic, token.Error())
	}
	return nil
}
