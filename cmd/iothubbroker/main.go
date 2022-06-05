package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/filariow/thsock/pkg/iothubmqtt"
)

const SockAddr = "unix:/tmp/th.socket"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ihc, err := setupMQTTClient()
	if err != nil {
		return err
	}

	configureHTTPServer(ihc)

	p := "8080"
	log.Printf("Serving on port: %s", p)
	return http.ListenAndServe(":"+p, nil)
}

func setupMQTTClient() (iothubmqtt.MQTTClient, error) {
	cfg, err := iothubmqtt.BuildConfigFromEnv("IOT_")
	if err != nil {
		return nil, fmt.Errorf("error building configuration for MQTT Client: %w", err)
	}

	ihc := iothubmqtt.NewMQTTClient(cfg)

	ihc.Configure()
	ihc.OnConnect(func(c mqtt.Client) { log.Println("MQTT Client connect") })

	if err := ihc.Connect(); err != nil {
		return nil, err
	}

	return ihc, nil
}

func configureHTTPServer(ihc iothubmqtt.MQTTClient) {
	log.Println("Configuring HTTP server")

	cfg := ihc.Configuration()
	t := fmt.Sprintf("devices/%s/messages/events/$.ct=application%%2Fjson&$.ce=utf-8", cfg.ClientID)
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			m := fmt.Sprintf("Expected method %s, found: %s", http.MethodPost, r.Method)
			log.Println(m)

			w.Write([]byte(m))
			w.WriteHeader(405)
			return
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			m := fmt.Sprintf("error reading request body: %s", err)
			log.Println(m)

			w.Write([]byte(m))
			w.WriteHeader(400)
			return
		}

		if err := sendMessageToIoT(ihc, t, string(b)); err != nil {
			m := fmt.Sprintf("error sending message to IoT Hub: %s", err)
			log.Println(m)

			w.Write([]byte(m))
			w.WriteHeader(400)
			return
		}

		w.WriteHeader(200)
	})
}

func sendMessageToIoT(ihc iothubmqtt.MQTTClient, topic, msg string) error {
	if err := ihc.Publish(topic, msg); err != nil {
		return err
	}

	log.Printf("Message '%s' sent to topic '%s'", msg, topic)
	return nil
}
