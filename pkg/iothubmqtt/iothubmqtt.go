package iothubmqtt

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient interface {
	Configure()
	OnConnect(func(mqtt.Client))
	OnConnectionLost(func(mqtt.Client, error))
	Connect() error
	IsConnected() bool
	Configuration() Config

	Publish(topic, message string) error
	Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token
}

func NewMQTTClient(config *Config) MQTTClient {
	return &mqttClient{config: *config}
}

type mqttClient struct {
	config            Config
	client            mqtt.Client
	mqttClientOptions *mqtt.ClientOptions
}

func (c *mqttClient) IsConnected() bool {
	return c.client != nil
}

func (c *mqttClient) Configuration() Config {
	return c.config
}

func (c *mqttClient) Configure() {
	cfg := c.config
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", cfg.Broker, cfg.Port))
	opts.SetClientID(cfg.ClientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetProtocolVersion(4)

	opts.OnConnect = func(client mqtt.Client) { log.Printf("MQTT Client connected to '%s'", cfg.Broker) }
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("MQTT Client lost connection to '%s': '%s'", cfg.Broker, err)
	}
	c.mqttClientOptions = opts
}

func (c *mqttClient) OnConnect(f func(mqtt.Client)) {
	c.mqttClientOptions.OnConnect = f
}

func (c *mqttClient) OnConnectionLost(f func(mqtt.Client, error)) {
	c.mqttClientOptions.OnConnectionLost = f
}

func (c *mqttClient) Connect() error {
	client := mqtt.NewClient(c.mqttClientOptions)
	// throw an error if the connection isn't successfull
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Failed to connect: %w", token.Error())
	}

	c.client = client
	return nil
}

func (c *mqttClient) Publish(topic, message string) error {
	token := c.client.Publish(topic, 0, false, message)
	<-token.Done()
	if token.Error() != nil {
		return fmt.Errorf("Failed to publish to topic '%s': %w", topic, token.Error())
	}
	return nil
}

func (c *mqttClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	return c.client.Subscribe(topic, qos, callback)
}
