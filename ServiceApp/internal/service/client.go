package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Product struct {
	Name  string `json:"name"`
	Price uint64 `json:"price"`
}

type ClientIface interface {
	GetProduct(ctx context.Context, sku int64) (*Product, error)
}

type ProductClient struct {
	baseURL string
	token   string
	client  *http.Client
	retries uint64
	delay   time.Duration
}

func NewClient(baseURL, token string, retries uint64, delay time.Duration) *ProductClient {
	return &ProductClient{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 5 * time.Second},
		retries: retries,
		delay:   delay,
	}
}

func (c *ProductClient) GetProduct(ctx context.Context, sku uint64) (*Product, error) {
	reqBody := map[string]any{"token": c.token, "sku": sku}
	b, _ := json.Marshal(reqBody)

	var lastErr error
	for i := 0; i < int(c.retries); i++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/get_product", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == 420 || resp.StatusCode == 429 {
				lastErr = fmt.Errorf("rate limited: %d", resp.StatusCode)
			} else if resp.StatusCode == http.StatusNotFound {
				return nil, ErrProductNotFound
			} else if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("bad status: %d", resp.StatusCode)
			} else {
				var p Product
				if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
					return nil, err
				}
				return &p, nil
			}
		}
		time.Sleep(c.delay)
	}

	if lastErr == nil {
		lastErr = errors.New("unknown product client error")
	}

	return nil, lastErr
}
