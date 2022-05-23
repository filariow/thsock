package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"context"

	"github.com/filariow/thsock/pkg/ihbclient"
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

	ctx := context.Background()
	if err := sendMessageToIoT(ctx, b); err != nil {
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

func sendMessageToIoT(ctx context.Context, data []byte) error {
	a := os.Getenv("IOT_ADDRESS")
	c := ihbclient.NewClient(a)

	r, err := c.SendEvent(ctx, data)
	if err != nil {
		return err
	}
	if r.StatusCode > 299 {
		return fmt.Errorf("Error sending message to IoTHub Broker at '%s'. Response: %v", a, *r)
	}

	return nil
}
