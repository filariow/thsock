package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Shopify/sarama"
)

const (
	envBrokerUrlsKey     = "WEATHER_BROKER_URLS"
	envBrokerUrlsDefault = "localhost:9092"

	envTopicKey     = "WEATHER_TOPIC"
	envTopicDefault = "temp-hum"
)

func main() {
	getEnvOrDefault := func(key, value string) string {
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
		return value
	}
	topic := getEnvOrDefault(envTopicKey, envTopicDefault)
	uc := getEnvOrDefault(envBrokerUrlsKey, envBrokerUrlsDefault)
	uu := strings.Split(strings.TrimRight(uc, ";"), ";")
	worker, err := connectConsumer(uu)
	if err != nil {
		log.Panic(err)
	}

	// Calling ConsumePartition. It will open one connection per broker
	// and share it for all partitions that live on it.
	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Consumer started ")
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	// Count how many message processed
	msgCount := 0

	// Get signal for finish
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println(err)
			case msg := <-consumer.Messages():
				msgCount++
				fmt.Printf("Received message Count %d: | Topic(%s) | Message(%s) \n", msgCount, string(msg.Topic), string(msg.Value))
			case <-sigchan:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	fmt.Println("Processed", msgCount, "messages")

	if err := worker.Close(); err != nil {
		log.Panic(err)
	}

}

func connectConsumer(brokersUrl []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Create new consumer
	conn, err := sarama.NewConsumer(brokersUrl, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
