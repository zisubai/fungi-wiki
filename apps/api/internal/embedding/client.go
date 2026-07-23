package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var ErrDisabled = errors.New("embedding provider is not configured")

type Provider interface {
	Embed(context.Context, []string) ([][]float32, error)
	Model() string
	Enabled() bool
}

type Client struct {
	url, key, model string
	http            *http.Client
}

func NewClient(url, key, model string) *Client {
	return &Client{url: strings.TrimSpace(url), key: key, model: strings.TrimSpace(model), http: &http.Client{Timeout: 20 * time.Second}}
}
func (c *Client) Model() string { return c.model }
func (c *Client) Enabled() bool { return c.url != "" && c.model != "" }

func (c *Client) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	if !c.Enabled() {
		return nil, ErrDisabled
	}
	body, err := json.Marshal(map[string]any{"model": c.model, "input": inputs})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.key != "" {
		req.Header.Set("Authorization", "Bearer "+c.key)
	}
	response, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding request returned %s", response.Status)
	}
	var result struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Data) != len(inputs) {
		return nil, fmt.Errorf("embedding response count mismatch")
	}
	output := make([][]float32, len(inputs))
	for _, item := range result.Data {
		if item.Index < 0 || item.Index >= len(output) || len(item.Embedding) == 0 {
			return nil, fmt.Errorf("invalid embedding response")
		}
		output[item.Index] = item.Embedding
	}
	return output, nil
}
