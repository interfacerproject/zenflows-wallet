package payout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Client posts issuance decisions to the Zenflows PDP.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClientFromEnv reads ZENFLOWS_PDP_URL. Returned even when empty; Decide
// then errors and InterpretPayouts fails closed.
func NewClientFromEnv() *Client {
	return &Client{BaseURL: os.Getenv("ZENFLOWS_PDP_URL"), HTTP: &http.Client{Timeout: 5 * time.Second}}
}

// Decide POSTs req to <BaseURL>/authz/decide and decodes the decision. A
// missing base URL, transport failure, non-2xx, or undecodable body errors.
func (c *Client) Decide(ctx context.Context, req Request) (Response, error) {
	var zero Response
	if c.BaseURL == "" {
		return zero, fmt.Errorf("ZENFLOWS_PDP_URL not configured")
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return zero, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/authz/decide", bytes.NewReader(buf))
	if err != nil {
		return zero, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(httpReq)
	if err != nil {
		return zero, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return zero, fmt.Errorf("pdp returned status %d", res.StatusCode)
	}
	var out Response
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return zero, err
	}
	return out, nil
}
