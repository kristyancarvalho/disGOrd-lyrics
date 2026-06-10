package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultEndpoint = "https://discord.com/api/v9/users/@me/settings"

type Client struct {
	token    string
	endpoint string
	client   *http.Client
}

type settingsPayload struct {
	CustomStatus *customStatus `json:"custom_status"`
}

type customStatus struct {
	Text string `json:"text"`
}

func New(token string) *Client {
	return NewWithURL(token, defaultEndpoint, &http.Client{Timeout: 5 * time.Second})
}

func NewWithURL(token, endpoint string, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &Client{token: token, endpoint: endpoint, client: client}
}

func Payload(text string) ([]byte, error) {
	payload := settingsPayload{}
	if text != "" {
		payload.CustomStatus = &customStatus{Text: text}
	}
	return json.Marshal(payload)
}

func (client *Client) Update(ctx context.Context, text string) error {
	return client.send(ctx, text)
}

func (client *Client) Clear(ctx context.Context) error {
	return client.send(ctx, "")
}

func (client *Client) send(ctx context.Context, text string) error {
	body, err := Payload(text)
	if err != nil {
		return fmt.Errorf("encode Discord status request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPatch, client.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("prepare Discord status request: %w", err)
	}
	request.Header.Set("Authorization", client.token)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.client.Do(request)
	if err != nil {
		return fmt.Errorf("send Discord status request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("send Discord status request: unexpected HTTP status %d", response.StatusCode)
	}

	return nil
}
