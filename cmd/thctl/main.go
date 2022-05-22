package main

import (
	"encoding/json"
	"fmt"
	"log"

	"context"

	"github.com/filariow/thsock/pkg/iothubmqtt"
	"github.com/filariow/thsock/pkg/thprotos"
	"google.golang.org/grpc"
)

const SockAddr = "unix:/tmp/th.socket"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	b, err := readSensor()
	if err != nil {
		return fmt.Errorf("error reading data from sensor: %w", err)
	}
	log.Printf("Read data from sensor: '%s'\n", b)

	m := string(b)
	if err := sendMessageToIoT(m); err != nil {
		return fmt.Errorf("error sending message to IoT Hub: %w", err)
	}

	return nil
}

func readSensor() ([]byte, error) {
	conn, err := grpc.Dial(SockAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	cli := thprotos.NewTempHumSvcClient(conn)
	ctx := context.Background()
	m, err := cli.ReadTempHum(ctx, &thprotos.ReadTempHumRequest{})
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("error marshaling ReadTempHumResponse (%v) as JSON: %w", m, err)
	}
	return b, nil
}

func sendMessageToIoT(msg string) error {
	cfg, err := iothubmqtt.BuildConfigFromEnv("IOT_")
	if err != nil {
		return fmt.Errorf("error building configuration for MQTT Client: %w", err)
	}

	t := fmt.Sprintf("devices/%s/messages/events/", cfg.ClientID)
	ihc := iothubmqtt.NewMQTTClient(cfg)
	if err := ihc.Publish(t, msg); err != nil {
		return err
	}

	log.Printf("Message '%s' sent to topic '%s'", msg, t)
	return nil
}
