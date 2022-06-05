package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"context"

	"github.com/filariow/thsock/pkg/ihbclient"
	"github.com/filariow/thsock/pkg/thprotos"
	"google.golang.org/grpc"
)

const SockAddr = "unix:/tmp/th.socket"

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
	mux := http.NewServeMux()
	mux.HandleFunc("/setDelay", func(w http.ResponseWriter, r *http.Request) {
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

		delay = data.Delay
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(`{"message": "delay set to %d"}`, delay)))
		return
	})
	srv := http.Server{Addr: "80", Handler: mux}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	srv.ListenAndServe()
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

		st := 5
		fmt.Printf("Sleeping %d seconds...", st)
		time.Sleep(time.Duration(st) * time.Second)
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
