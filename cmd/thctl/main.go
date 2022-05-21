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
	conn, err := grpc.Dial(SockAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	cli := thprotos.NewTempHumSvcClient(conn)
	ctx := context.Background()
	m, err := cli.ReadTempHum(ctx, &thprotos.ReadTempHumRequest{})
	if err != nil {
		return err
	}

	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("error marshaling ReadTempHumResponse (%v) as JSON: %w", m, err)
	}
	fmt.Printf("%s\n", b)

	if err := sendMessageToIoT(string(b)); err != nil {
		return fmt.Errorf("error sending message to IoT Hub: %w", err)
	}
	return nil
}

func sendMessageToIoT(msg string) error {
	cfg, err := iothubmqtt.BuildConfigFromEnv("IOT_")
	if err != nil {
		return fmt.Errorf("error building configuration for MQTT Client: %w", err)
	}

	ihc := iothubmqtt.NewMQTTClient(cfg)
	t := "garden/temphum"
	if err := ihc.Publish(t, msg); err != nil {
		return err
	}
	log.Printf("Message '%s' sent to topic '%s'", msg, t)
	return nil

}
