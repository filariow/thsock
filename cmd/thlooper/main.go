package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"context"

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

type SetDelayData struct {
	Delay int `json:"delay"`
}

func startHTTPServer(ctx context.Context) {
	http.HandleFunc("/setDelay", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Processing request to change sampling delay")
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			w.Write([]byte(fmt.Sprintf(`{"error": "method '%s' not supported, use POST"}`, r.Method)))
			return
		}

		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf(`{"error": "request body not valid: %s"}`, err.Error())))
			return
		}

		var data SetDelayData
		if err := json.Unmarshal(bs, &data); err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf(`{"error": "request body not valid: %s"}`, err.Error())))
			return
		}

		if minDelay := 1000; data.Delay <= minDelay {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf(`{"error": "delay must be bigger than %d"}`, minDelay)))
			return
		}

		log.Printf("Setting delay time to %d ms", data.Delay)
		delayMux.Lock()
		delay = data.Delay
		delayMux.Unlock()

		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(`{"message": "delay set to %d"}`, delay)))
		return
	})
	http.ListenAndServe(":8080", nil)
}

func run() error {
	a := os.Getenv("IOT_ADDRESS")
	c, err := ihbclient.NewClient(a)
	if err != nil {
		return fmt.Errorf("error creating client for IoT Hub Broker: %w", err)
	}

	ctx := context.Background()
	go startHTTPServer(ctx)

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
