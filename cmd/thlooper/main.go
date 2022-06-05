package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"context"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/filariow/thsock/pkg/ihbclient"
	"github.com/filariow/thsock/pkg/iothubmqtt"
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

func run() error {
	a := os.Getenv("IOT_ADDRESS")
	c, err := ihbclient.NewClient(a)
	if err != nil {
		return fmt.Errorf("error creating client for IoT Hub Broker: %w", err)
	}

	cfg, err := loadMQTTClientConfig()
	if err != nil {
		return err
	}

	if _, err := setupMQTTClient(cfg); err != nil {
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

			t, err := parseTopic(m.Topic())
			if err != nil {
				log.Printf("error parsing topic '%s': %s", m.Topic(), err)
				return
			}

			var data SetDelayData
			if err := json.Unmarshal(m.Payload(), &data); err != nil {
				respondToDirectMethodExecution(c, t.rid, 400,
					fmt.Sprintf(`{"error": "error parsing payload: %s"}`, err))
				return
			}

			if minDelay := 1000; data.Delay <= minDelay {
				respondToDirectMethodExecution(c, t.rid, 400,
					fmt.Sprintf(`{"error": "delay must be bigger than %d"}`, minDelay))
				return
			}

			log.Printf("Setting delay time to %d ms", data.Delay)
			delayMux.Lock()
			delay = data.Delay
			delayMux.Unlock()

			msg := fmt.Sprintf(`{"message": "delay set to %d"}`, delay)
			respondToDirectMethodExecution(c, t.rid, 200, msg)
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
				log.Printf("Setting delay time to %d ms", *data.THLooper.Delay)
				delayMux.Lock()
				delay = *data.THLooper.Delay
				delayMux.Unlock()
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
