package ihbclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path"
)

type IHBClient interface {
	SendEvent(ctx context.Context, data []byte) (*http.Response, error)
}

func NewClient(address string) IHBClient {
	return &ihbClient{address: address}
}

type ihbClient struct {
	address string
}

type thmodel struct {
	Temperature float64 `json:"Temperature"`
	Humidity    float64 `json:"Humidity"`
}

func (c *ihbClient) SendEvent(ctx context.Context, data []byte) (*http.Response, error) {
	u := path.Join(c.address, "event")

	br := bytes.NewReader(data)
	rq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, br)
	if err != nil {
		return nil, fmt.Errorf("error building new POST request with context with address '%s': %w", u, err)
	}

	r, err := http.DefaultClient.Do(rq)
	if err != nil {
		return nil, fmt.Errorf("error sending event to '%s': %w", c.address, err)
	}

	return r, nil
}
