package main

import (
	"encoding/json"
	"fmt"
	"log"

	"context"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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
		return fmt.Errorf("Error marshaling ReadTempHumResponse (%v) as JSON: %w", m, err)
	}
	fmt.Printf("%s\n", b)

	// sendSampleToIoTHub(string(b))
	return nil
}

// upon connection to the client, this is called
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

// this is called when the connection to the client is lost, it prints "Connection lost" and the corresponding error
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

func sendSampleToIoTHub(text string) error {
	var broker = "<host_name>"
	var port = 8883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", broker, port))
	opts.SetClientID("<client_name>")
	opts.SetUsername("<username>")
	opts.SetPassword("<password>")

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	// create the client using the options above

	client := mqtt.NewClient(opts)
	// throw an error if the connection isn't successfull
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Failed to connect: %w", token.Error())
	}

	token := client.Publish("topic/test", 0, false, text)
	<-token.Done()
	if token.Error() != nil {
		return fmt.Errorf("Failed to publish to topic: %w", token.Error())
	}
	return nil
}
