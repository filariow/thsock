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

	cfg, err := iothubmqtt.BuildConfigFromEnv("")
	if err != nil {
		return fmt.Errorf("error building configuration for IoTHub's MQTT Client: %w", err)
	}

	ihc := iothubmqtt.NewMQTTClient(cfg)
	err = ihc.Publish("garden/temphum", string(b))
	return err
}
