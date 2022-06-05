package thiotsub

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/filariow/thsock/pkg/iothubmqtt"
)

func SetupMQTTClient(setDelay func(delay int)) (iothubmqtt.MQTTClient, error) {
	cfg, err := iothubmqtt.BuildConfigFromEnv("IOT_")
	if err != nil {
		return nil, fmt.Errorf("error building configuration for MQTT Client: %w", err)
	}
	ihc := iothubmqtt.NewMQTTClient(cfg)

	ihc.Configure()
	ihc.OnConnect(buildOnConnectHandler(ihc, setDelay))

	if err := ihc.Connect(); err != nil {
		return nil, err
	}

	return ihc, nil
}

func buildOnConnectHandler(ihc iothubmqtt.MQTTClient, setDelay func(delay int)) func(c mqtt.Client) {
	return func(c mqtt.Client) {
		log.Println("MQTT Client connect")

		td := "$iothub/methods/POST/#"
		log.Printf("Subscribing to direct method topic: %s", td)
		tkn := ihc.Subscribe(td, 0, buildDirectMethodRequestHandler(setDelay))

		tpd := "$iothub/twin/PATCH/properties/desired/#"
		log.Printf("Subscribing to desired properties topic: %s", tpd)
		tkn2 := ihc.Subscribe(tpd, 0, buildDesiredPropertiesRequestHandler(setDelay))

		<-tkn.Done()
		if err := tkn.Error(); err != nil {
			log.Fatalln(err)
		}
		<-tkn2.Done()
		if err := tkn2.Error(); err != nil {
			log.Fatalln(err)
		}
	}
}

func buildDesiredPropertiesRequestHandler(setDelay func(delay int)) func(c mqtt.Client, m mqtt.Message) {
	return func(c mqtt.Client, m mqtt.Message) {
		log.Printf("processing message from topic '%s' for desired properties", m.Topic())

		if m.Payload() == nil {
			log.Println("message paylod for desired properties notification is empty, skipping.")
			return
		}

		log.Printf("desired properties payload: %s", m.Payload())
		var data SetDelayData
		if err := json.Unmarshal(m.Payload(), &data); err != nil {
			log.Printf("error unmarshaling desired properties json: %s", err)
			return
		}

		setDelay(data.Delay)
	}
}

func buildDirectMethodRequestHandler(setDelay func(delay int)) func(c mqtt.Client, m mqtt.Message) {
	return func(c mqtt.Client, m mqtt.Message) {
		log.Printf("Message received '%d' on topic %s: %s", m.MessageID(), m.Topic(), m.Payload())

		t, err := parseTopic(m.Topic())
		if err != nil {
			log.Printf("error parsing topic '%s': %s", m.Topic(), err)
			return
		}

		if t.method == "setDelay" {
			log.Println("Invoked set delay")

			var data SetDelayData
			if err := json.Unmarshal(m.Payload(), &data); err != nil {
				respondToDirectMethodExecution(c, t.rid, 400,
					fmt.Sprintf(`{"error": "error parsing payload: %s"}`, err))
				return
			}

			if minDelay := 1000; data.Delay < minDelay {
				respondToDirectMethodExecution(c, t.rid, 400,
					fmt.Sprintf(`{"error": "delay must be bigger than %d"}`, minDelay))
				return
			}

			log.Printf("Setting delay time to %d ms", data.Delay)
			setDelay(data.Delay)

			msg := fmt.Sprintf(`{"message": "delay set to %d"}`, data.Delay)
			respondToDirectMethodExecution(c, t.rid, 200, msg)
		}
	}
}

type SetDelayData struct {
	Delay int `json:"delay"`
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
	method string
	rid    string
}

func parseTopic(topic string) (*directMethodData, error) {
	rg := `^\$iothub\/methods\/POST\/([A-z]+)\/\?\$rid=([0-9]*)$`
	re, err := regexp.Compile(rg)
	if err != nil {
		return nil, err
	}
	mm := re.FindSubmatch([]byte(topic))
	if exp := 3; len(mm) != exp {
		return nil, fmt.Errorf("expected %d matches from topic '%s' and regexp '%s', found %d", exp, topic, rg, len(mm))
	}
	meth, rid := mm[1], mm[2]
	return &directMethodData{
		method: string(meth),
		rid:    string(rid),
	}, nil
}
