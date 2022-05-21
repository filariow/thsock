package main

import (
	"fmt"
	"log"

	"github.com/filariow/thsock/pkg/iothubmqtt"
)

const SockAddr = "unix:/tmp/th.socket"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := iothubmqtt.BuildConfigFromEnv("IOT_")
	if err != nil {
		return fmt.Errorf("error building configuration for IoTHub's MQTT Client: %w", err)
	}

	fmt.Printf("%+v\n", cfg)
	ihc := iothubmqtt.NewMQTTClient(cfg)
	err = ihc.Publish("garden/temphum", "hello")
	return err
}
