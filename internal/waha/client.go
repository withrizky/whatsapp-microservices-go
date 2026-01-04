package waha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"whatsapp_microservices/internal/model"
)

type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func NewClient(url, key string) *Client {
	return &Client{
		BaseURL: url,
		APIKey:  key,
		HTTP:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) SendText(job model.WaPayload, session string) error {
	url := fmt.Sprintf("%s/api/sendText", c.BaseURL)
	payload := map[string]string{
		"chatId":  job.To + "@c.us",
		"text":    job.Message,
		"session": session,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return nil
}
