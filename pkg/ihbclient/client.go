package ihbclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type IHBClient interface {
	SendEvent(ctx context.Context, data []byte) (*http.Response, error)
}

func NewClient(address string) (IHBClient, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	return &ihbClient{address: *u}, err
}

type ihbClient struct {
	address url.URL
}

type thmodel struct {
	Temperature float64 `json:"Temperature"`
	Humidity    float64 `json:"Humidity"`
}

func (c *ihbClient) SendEvent(ctx context.Context, data []byte) (*http.Response, error) {
	a := c.address
	a.Path = "/event"

	u := a.String()
	br := bytes.NewReader(data)
	rq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, br)
	if err != nil {
		return nil, fmt.Errorf("error building new POST request with context with address '%s': %w", u, err)
	}

	r, err := http.DefaultClient.Do(rq)
	if err != nil {
		return nil, fmt.Errorf("error sending event to '%s': %w", u, err)
	}

	return r, nil
}
