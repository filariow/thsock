package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"context"

	"github.com/filariow/thsock/internal/thiotsub"
	"github.com/filariow/thsock/pkg/ihbclient"
	"github.com/filariow/thsock/pkg/thprotos"
	"google.golang.org/grpc"
)

const SockAddr = "unix:/tmp/th.socket"

var delayMux sync.RWMutex
var delay = 5000

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	a := os.Getenv("IOT_ADDRESS")
	c, err := ihbclient.NewClient(a)
	if err != nil {
		return fmt.Errorf("error creating client for IoT Hub Broker: %w", err)
	}

	if _, err := thiotsub.SetupMQTTClient(setDelay); err != nil {
		return err
	}

	for {
		b, err := readSensor()
		if err != nil {
			return fmt.Errorf("error reading data from sensor: %w", err)
		}
		log.Printf("Read data from sensor: '%s'\n", b)

		ctx := context.Background()
		if err := sendMessageToIoT(ctx, c, b); err != nil {
			return fmt.Errorf("error sending message to IoT Hub: %w", err)
		}

		delayMux.RLock()
		d := delay
		delayMux.RUnlock()

		log.Printf("Sleeping %d milliseconds...", d)
		time.Sleep(time.Duration(d) * time.Millisecond)
	}
}

func setDelay(d int) {
	delayMux.Lock()
	delay = d
	delayMux.Unlock()
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

func sendMessageToIoT(ctx context.Context, c ihbclient.IHBClient, data []byte) error {
	a := c.Address()
	log.Printf("Sending data to IoT Hub broker at '%s'", a)

	r, err := c.SendEvent(ctx, data)
	if err != nil {
		return err
	}
	if r.StatusCode < 200 || r.StatusCode > 299 {
		return fmt.Errorf("Error sending message to IoTHub Broker at '%s'. Response: %v", a, *r)
	}

	log.Println("Data sent successfully")
	return nil
}
