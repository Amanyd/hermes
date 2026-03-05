package slack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// Happy-path test for slack integration
func TestExecute_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
	}))
	defer srv.Close()

	sender := New()
	sender.client = srv.Client()
	cfg := map[string]any{
		"webhook_url": srv.URL,
	}

	_, err := sender.Execute(context.Background(), cfg, []byte(`{"event":"test"}`), nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// Verifies that a missing webhook_url is rejected immediately
func TestExecute_MissingURL(t *testing.T) {
	sender := New()
	_, err := sender.Execute(context.Background(), map[string]any{}, []byte(`{}`), nil)
	if err == nil {
		t.Error("expected error for missing webhook_url")
	}
}

// Verifies that a 4xx (except 429) error fails immediately without retrying
func TestExecute_NonRetryableError(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	sender := New()
	sender.client = srv.Client()
	_, err := sender.Execute(context.Background(), map[string]any{"webhook_url": srv.URL}, []byte(`{}`), nil)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}

	if got := callCount.Load(); got != 1 {
		t.Errorf("expected 1 call (no retries), got: %d", got)
	}
}

// Verifies that 429 triggers retries
func TestExecute_RetryOn429(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	sender := New()
	sender.client = srv.Client()
	_, err := sender.Execute(context.Background(), map[string]any{"webhook_url": srv.URL}, []byte(`{}`), nil)
	if err == nil {
		t.Fatal("expected error after retries")
	}
	if got := callCount.Load(); got != 3 {
		t.Errorf("expected 3 calls (with retries), got: %d", got)
	}
}

// Verifies that 5xx responses trigger retries
func TestExecute_RetryOnServerError(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer srv.Close()
	sender := New()
	sender.client = srv.Client()
	_, err := sender.Execute(context.Background(), map[string]any{"webhook_url": srv.URL}, []byte(`{}`), nil)
	if err == nil {
		t.Fatal("expected error after retries")
	}
	if got := callCount.Load(); got != 3 {
		t.Errorf("expected 3 calls (with retries), got %d", got)
	}
}

func TestExecute_CustomTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	sender := New()
	sender.client = srv.Client()
	cfg := map[string]any{
		"webhook_url":      srv.URL,
		"message_template": "Custom notification: yo pierre you wanna come out here?",
	}
	_, err := sender.Execute(context.Background(), cfg, []byte(`{}`), nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
