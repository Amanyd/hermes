package httpreq

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var allowedMethods = map[string]bool{
	http.MethodGet:    true,
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodDelete: true,
	http.MethodPatch:  true,
}

type Executor struct {
	client *http.Client
}

func New() *Executor {
	return &Executor{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (e *Executor) Execute(ctx context.Context, cfg map[string]any, payload []byte) error {
	url, _ := cfg["url"].(string)
	if url == "" {
		return fmt.Errorf("missing url in http_request config")
	}
	method, _ := cfg["method"].(string)
	method = strings.ToUpper(method)
	if method == "" {
		method = http.MethodPost
	}
	if !allowedMethods[method] {
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}
	var body io.Reader
	if method != http.MethodGet {
		bodyTemplate, _ := cfg["body_template"].(string)
		if bodyTemplate != "" {
			body = strings.NewReader(bodyTemplate)
		} else {
			body = bytes.NewReader(payload)
		}
	}
	var lastErr error
	for attempt := range 3 {
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		if method != http.MethodGet {
			req.Header.Set("Content-Type", "application/json")
		}

		if headers, ok := cfg["headers"].(map[string]any); ok {
			for k, v := range headers {
				if s, ok := v.(string); ok {
					req.Header.Set(k, s)
				}
			}
		}

		resp, doErr := e.client.Do(req)
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
			lastErr = fmt.Errorf("http_request returned %d", resp.StatusCode)
			time.Sleep(time.Duration(300*(attempt)+1) * time.Millisecond)
			continue
		}
		return fmt.Errorf("http_request returned non-retryable status %d", resp.StatusCode)
	}
	return fmt.Errorf("http_request failed after retries: %w", lastErr)
}
