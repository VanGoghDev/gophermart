package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
)

// Client общается с внешним сервисом Accrual.
type Client struct {
	host   string
	client http.Client
}

var (
	ErrToManyRequests = errors.New("too many requests")
)

func New(client http.Client, accrlHost string) *Client {
	return &Client{
		client: client,
		host:   accrlHost,
	}
}

func (c *Client) GetAccrual(ctx context.Context, orderNum string) (order models.Accrual, timeout time.Duration, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/orders/%s", c.host, orderNum),
		http.NoBody,
	)
	if err != nil {
		return models.Accrual{}, 0, fmt.Errorf("failed to init request: %w", err)
	}

	r, err := c.client.Do(req)
	if err != nil {
		return models.Accrual{}, 0, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		errc := r.Body.Close()
		if errc != nil {
			err = errc
		}
	}()
	if r.StatusCode == http.StatusNoContent {
		return models.Accrual{}, 0, errors.New("order is not registered")
	}

	if r.StatusCode == http.StatusTooManyRequests {
		return models.Accrual{}, time.Second, fmt.Errorf("%w: accrual response with 429 status code", ErrToManyRequests)
	}

	var accrl models.Accrual
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&accrl)
	if err != nil {
		return models.Accrual{}, 0, fmt.Errorf("failed to decode json: %w", err)
	}

	return accrl, 0, nil
}
