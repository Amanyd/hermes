package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const resendURL = "http://api.resend.com/emails"

type Sender struct {
	client *http.Client
}

func New() *Sender {
	return &Sender{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Sender) Execute(ctx context.Context, cfg map[string]any, payload []byte) error {
	apiKey, _ := cfg["api_key"].(string)
	if apiKey == "" {
		return fmt.Errorf("missing api_key in email_send config")
	}
	from, _ := cfg["from"].(string)
	if from == "" {
		return fmt.Errorf("missing from in email_send config")
	}
	to, _ := cfg["to"].(string)
	if to == "" {
		return fmt.Errorf("missing to in email_send config")
	}

	subject, _ := cfg["subject"].(string)
	if subject == "" {
		subject = "Hermes Notification"
	}

	body, _ := cfg["body"].(string)
	if body == "" {
		body = fmt.Sprintf("Payload received:\n%s", string(payload))
	}
	emailBody := map[string]string{
		"from":    from,
		"to":      to,
		"subject": subject,
		"text":    body,
	}
	bodyJSON, err := json.Marshal(emailBody)
	if err != nil {
		return fmt.Errorf("marshal email body:  %w", err)
	}

	var lastErr error
	for attempt := range 3 {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, resendURL, bytes.NewBuffer(bodyJSON))
		if reqErr != nil {
			return fmt.Errorf("build request:  %w", reqErr)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, doErr := s.client.Do(req)
		if doErr != nil {
			lastErr = doErr
			time.Sleep(time.Duration(300*(attempt+1)) * time.Millisecond)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("resend returned %d", resp.StatusCode)
			time.Sleep(time.Duration(300*(attempt+1)) * time.Millisecond)
			continue
		}
		return fmt.Errorf("resend returned non-retryable status %d", resp.StatusCode)
	}
	return fmt.Errorf("email_send failed after retries: %w", lastErr)
}
