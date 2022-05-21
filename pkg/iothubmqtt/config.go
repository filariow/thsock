package iothubmqtt

import (
	"os"
	"strconv"
)

type Config struct {
	Broker   string
	Port     int
	ClientID string
	Username string
	Password string
}

func BuildConfigFromEnv(prefix string) (*Config, error) {

	b := os.Getenv(prefix + "BROKER")
	p := os.Getenv(prefix + "PORT")
	pi, err := strconv.Atoi(p)
	if err != nil {
		return nil, err
	}
	c := os.Getenv(prefix + "CLIENTID")
	u := os.Getenv(prefix + "USERNAME")
	pd := os.Getenv(prefix + "PASSWORD")

	return &Config{
		Broker:   b,
		Port:     pi,
		ClientID: c,
		Username: u,
		Password: pd,
	}, nil
}
