package httpreq

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// Verifies a successful POST request
func TestExecute_PostSuccess(t *testing.T) {
	var receivedBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got: %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	executor := New()
	executor.client = srv.Client()
	cfg := map[string]any{
		"url":    srv.URL,
		"method": "POST",
	}
	payload := []byte(`{"data":"value"`)
	err := executor.Execute(context.Background(), cfg, payload)
	if err != nil {
		t.Fatalf("expected no error, go: %v", err)
	}
	if receivedBody != `{"data":"value"}` {
		t.Errorf("unexpected body: %s", receivedBody)
	}
}

// Verifies that an omitted method defaults to POST
func TestExecute_DefaultMethodIsPost(t *testing.T) {
	var receivedMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	executor := New()
	executor.client = srv.Client()
	err := executor.Execute(context.Background(), map[string]any{"url": srv.URL}, []byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got: %s", receivedMethod)
	}
}

// Ensures a missing URL returns an error immediately
func TestExecute_MissingURL(t *testing.T) {
	executor := New()
	err := executor.Execute(context.Background(), map[string]any{}, []byte(`{}`))
	if err == nil {
		t.Error("expected error for missing url")
	}
}

// Verifies that unsupported HTTP methods are rejected
func TestExecute_UnsupportedMethod(t *testing.T) {
	executor := New()
	err := executor.Execute(context.Background(), map[string]any{
		"url":    "http://blahblah.com",
		"method": "OPTIONS",
	}, []byte(`{}`))
	if err == nil {
		t.Error("expected error for unsupported method")
	}
}

// Verifies that custom headers from config are sent
func TestExecute_CustomHeaders(t *testing.T) {
	var capturedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("X-Api-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	executor := New()
	executor.client = srv.Client()
	cfg := map[string]any{
		"url":     srv.URL,
		"method":  "POST",
		"headers": map[string]any{"X-Api-Key": "secret-123"},
	}
	err := executor.Execute(context.Background(), cfg, []byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedAuth != "secret-123" {
		t.Errorf("expected header 'secret-123', got %q", capturedAuth)
	}
}

// Verifies 5xx retries
func TestExecute_RetryOnServerError(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()
	executor := New()
	executor.client = srv.Client()
	err := executor.Execute(context.Background(), map[string]any{"url": srv.URL}, []byte(`{}`))
	if err == nil {
		t.Fatal("expected error after retries")
	}
	if got := callCount.Load(); got != 3 {
		t.Errorf("expected 3 calls, got: %d", got)
	}
}
