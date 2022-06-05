package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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
	cfg, err := loadMQTTClientConfig()
	if err != nil {
		return err
	}

	ihc, err := setupMQTTClient(cfg)
	if err != nil {
		return err
	}

	t := fmt.Sprintf("devices/%s/messages/events/$.ct=application%%2Fjson&$.ce=utf-8", cfg.ClientID)

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

func loadMQTTClientConfig() (*iothubmqtt.Config, error) {
	cfg, err := iothubmqtt.BuildConfigFromEnv("IOT_")
	if err != nil {
		return nil, fmt.Errorf("error building configuration for MQTT Client: %w", err)
	}
	return cfg, nil
}

func setupMQTTClient(cfg *iothubmqtt.Config) (iothubmqtt.MQTTClient, error) {
	ihc := iothubmqtt.NewMQTTClient(cfg)

	ihc.Configure()
	ihc.OnConnect(func(c mqtt.Client) {
		log.Println("MQTT Client connect")

		td := "$iothub/methods/POST/#"
		log.Printf("Subscribing to direct method topic: %s", td)
		tkn := ihc.Subscribe(td, 0, func(c mqtt.Client, m mqtt.Message) {
			log.Printf("Message received '%d' on topic %s: %s", m.MessageID(), m.Topic(), m.Payload())

			rid := strings.Split(m.Topic(), "=")[1]
			tr := fmt.Sprintf("$iothub/methods/res/200/?$rid=%s", rid)
			log.Printf("Responding to message %d on topic '%s'", m.MessageID(), tr)
			st := c.Publish(tr, 0, false, `{"status":"ok"}`)
			<-st.Done()
			if err := st.Error(); err != nil {
				log.Println(err)
				return
			}
			log.Printf("Response sent on topic %s", tr)
		})
		<-tkn.Done()
		if err := tkn.Error(); err != nil {
			log.Fatalln(err)
		}
	})
	if err := ihc.Connect(); err != nil {
		return nil, err
	}

	return ihc, nil
}

func sendMessageToIoT(ihc iothubmqtt.MQTTClient, topic, msg string) error {
	if err := ihc.Publish(topic, msg); err != nil {
		return err
	}

	log.Printf("Message '%s' sent to topic '%s'", msg, topic)
	return nil
}
