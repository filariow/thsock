package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

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
		return fmt.Errorf("error building configuration for MQTT Client: %w", err)
	}

	t := fmt.Sprintf("devices/%s/messages/events/", cfg.ClientID)
	ihc := iothubmqtt.NewMQTTClient(cfg)

	log.Println("Configuring HTTP server")
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

	p := "8080"
	log.Printf("Serving on port: %s", p)
	return http.ListenAndServe(":"+p, nil)
}

func sendMessageToIoT(ihc iothubmqtt.MQTTClient, topic, msg string) error {
	if err := ihc.Publish(topic, msg); err != nil {
		return err
	}

	log.Printf("Message '%s' sent to topic '%s'", msg, topic)
	return nil
}
