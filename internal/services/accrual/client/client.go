package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
)

// Client общается с внешним сервисом Accrual.
type Client struct {
	host   string
	client http.Client
}

func New(client http.Client, accrlHost string) *Client {
	return &Client{
		client: client,
		host:   accrlHost,
	}
}

func (c *Client) GetAccrual(ctx context.Context, orderNum string) (order models.Accrual, err error) {
	const op = "services.accrual.client.GetAccrual"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/orders/%s", c.host, orderNum),
		http.NoBody,
	)
	if err != nil {
		return models.Accrual{}, fmt.Errorf("%s: %w", op, err)
	}

	r, err := c.client.Do(req)
	if err != nil {
		return models.Accrual{}, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		errc := r.Body.Close()
		if errc != nil {
			err = errc
		}
	}()

	if r.StatusCode == http.StatusNoContent {
		return models.Accrual{}, errors.New("order is not registered")
	}

	var accrl models.Accrual
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&accrl)
	if err != nil {
		return models.Accrual{}, fmt.Errorf("%s: %w", op, err)
	}

	if err != nil {
		return models.Accrual{}, fmt.Errorf("%s: %w", op, err)
	}

	return accrl, nil
}
