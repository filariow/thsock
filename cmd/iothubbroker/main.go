package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
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

			r, err := executeDirectMethod(m.Topic(), string(m.Payload()))
			if err != nil {
				log.Println("error executing direct method: %s", err)
				respondToDirectMethodExecution(c, r.rid, 500, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
				return
			}

			if r.payload == nil {
				respondToDirectMethodExecution(c, r.rid, r.statusCode, "")
				return
			}

			p, err := ioutil.ReadAll(r.payload)
			if err != nil {
				log.Println("error reading payload from service response to request with rid %d", r.rid)
				respondToDirectMethodExecution(c, r.rid, 500, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
				return
			}

			respondToDirectMethodExecution(c, r.rid, r.statusCode, string(p))
		})
		tkn2 := ihc.Subscribe("$iothub/twin/PATCH/properties/desired/#", 0, func(c mqtt.Client, m mqtt.Message) {
			log.Printf("processing message from topic '%s' for desired properties", m.Topic())

			if m.Payload() == nil {
				log.Println("message paylod for desired properties notification is empty, skipping.")
				return
			}

			log.Printf("desired properties payload: %s", m.Payload())
			var data ReportedPropertiesData
			if err := json.Unmarshal(m.Payload(), &data); err != nil {
				log.Printf("error unmarshaling desired properties json: %s", err)
				return
			}

			if data.THLooper != nil && data.THLooper.Delay != nil {
				r, err := http.Post(
					"http://thlooper.default.svc.cluster.local:8080/setDelay",
					"application/json",
					strings.NewReader(fmt.Sprintf(`{"delay": %d}`, data.THLooper.Delay)))
				if err != nil {
					log.Printf("error sending setDelay request to thlooper: %s", err)
					return
				}

				log.Printf("thlooper setDelay responded with %d", r.StatusCode)
			}
		})

		<-tkn.Done()
		if err := tkn.Error(); err != nil {
			log.Fatalln(err)
		}
		<-tkn2.Done()
		if err := tkn2.Error(); err != nil {
			log.Fatalln(err)
		}
	})
	if err := ihc.Connect(); err != nil {
		return nil, err
	}

	return ihc, nil
}

type ReportedPropertiesData struct {
	THLooper *struct {
		Delay *int `json:"delay"`
	} `json:"thlooper"`
}

func respondToDirectMethodExecution(c mqtt.Client, rid string, status int, payload string) {
	tr := fmt.Sprintf("$iothub/methods/res/%d/?$rid=%s", status, rid)
	log.Printf("Responding to message on topic '%s'", tr)
	st := c.Publish(tr, 0, false, payload)
	<-st.Done()
	if err := st.Error(); err != nil {
		log.Println(err)
	}
	log.Printf("Response sent on topic %s", tr)
}

func executeDirectMethod(topic, payload string) (*directMethodResponse, error) {
	m, err := parseTopic(topic)
	if err != nil {
		return nil, err
	}

	u := "http://" + path.Join(m.service, m.method)
	log.Printf("Sending post request to %s", u)
	r, err := http.Post(u, "application/json", strings.NewReader(payload))
	if err != nil {
		return &directMethodResponse{
			rid:        m.rid,
			statusCode: 500,
			payload:    io.NopCloser(strings.NewReader(fmt.Sprintf(`{"error": "%s"}`, err))),
		}, nil
	}

	return &directMethodResponse{
		rid:        m.rid,
		statusCode: r.StatusCode,
		payload:    r.Body,
	}, nil
}

type directMethodResponse struct {
	rid        string
	statusCode int
	payload    io.ReadCloser
}

type directMethodData struct {
	service string
	method  string
	rid     string
}

func parseTopic(topic string) (*directMethodData, error) {
	rg := `^\$iothub\/methods\/POST\/([A-z]+)_([A-z]+)\/\?\$rid=([0-9]*)$`
	re, err := regexp.Compile(rg)
	if err != nil {
		return nil, err
	}
	mm := re.FindSubmatch([]byte(topic))
	if exp := 4; len(mm) != exp {
		return nil, fmt.Errorf("expected %d matches from topic '%s' and regexp '%s', found %d", exp, topic, rg, len(mm))
	}
	svc, meth, rid := mm[1], mm[2], mm[3]
	return &directMethodData{
		service: string(svc) + ".default.svc.cluster.local:8080",
		method:  string(meth),
		rid:     string(rid),
	}, nil
}

func sendMessageToIoT(ihc iothubmqtt.MQTTClient, topic, msg string) error {
	if err := ihc.Publish(topic, msg); err != nil {
		return err
	}

	log.Printf("Message '%s' sent to topic '%s'", msg, topic)
	return nil
}
