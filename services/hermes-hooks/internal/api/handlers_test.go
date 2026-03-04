package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/logger"
	"github.com/go-chi/chi/v5"
)

// mockProducer captures the last published event and can simulate errors
type mockProducer struct {
	lastRelayID string
	lastEvent   ExecutionEvent
	publishErr  error
	callCount   int
}

func (m *mockProducer) Publish(relayID string, event ExecutionEvent) error {
	m.callCount++
	m.lastRelayID = relayID
	m.lastEvent = event
	return nil
}

// Creates a chi router wired to the handler so the URL params work
func newTestRouter(p EventProducer) *chi.Mux {
	testLogger := logger.New("hermes-hooks-test", "test", "debug")
	h := NewHandler(p, testLogger)
	r := chi.NewRouter()
	r.Post("/hooks/{relayID", h.HandleWebhook)
	return r
}

// Verifies the happy-path of the webhook handler
func TestHandlerWebhook_Success(t *testing.T) {
	mock := &mockProducer{}
	router := newTestRouter(mock)

	body := []byte(`{"test": "data"}`)
	req := httptest.NewRequest(http.MethodPost, "/hooks/relay-abc", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if mock.lastRelayID != "relay-abc" {
		t.Errorf("expected relayID='relay-abc', got %q", mock.lastRelayID)
	}

	if string(mock.lastEvent.Payload) != `{"test":"data"}` {
		t.Errorf("unexpected payload: %s", string(mock.lastEvent.Payload))
	}

	if mock.lastEvent.EventID == "" {
		t.Error("expected auto-generated event_id, got empty")
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if resp["status"] != "queued" {
		t.Errorf("expected status=queued, got %q", resp["status"])
	}
}

// Verifies that when the caller provides an X-Event-ID header, it's used instead of auto-generating one
func TestHandleWebhook_EventIDHeader(t *testing.T) {
	mock := &mockProducer{}
	router := newTestRouter(mock)
	req := httptest.NewRequest(http.MethodPost, "/hooks/relay-1", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-Event-ID", "custom-event-69")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if mock.lastEvent.EventID != "custom-event-69" {
		t.Errorf("expected event_id='custom-event-69', got %q", mock.lastEvent.EventID)
	}
}

// Verifies the query-param fallback when X-Event-ID header is missing
func TestHandleWebhook_EventIDFromQuery(t *testing.T) {
	mock := &mockProducer{}
	router := newTestRouter(mock)
	req := httptest.NewRequest(http.MethodPost, "/hooks/relay-1?event_id=query-event-69", bytes.NewReader([]byte(`{}`)))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if mock.lastEvent.EventID != "query-event-69" {
		t.Errorf("expected event_id='query-event-69', got %q", mock.lastEvent.EventID)
	}
}

// Verifies that when the queue returns an error,
// the webhook handler responds with 500 Internal Server Error.
func TestHandleWebhook_PublishError(t *testing.T) {
	mock := &mockProducer{
		publishErr: fmt.Errorf("NATS connection lost"),
	}
	router := newTestRouter(mock)

	req := httptest.NewRequest(http.MethodPost, "/hooks/relay-1", bytes.NewReader([]byte(`{}`)))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Verifies that an empty body is accepted
// The relay might not need a payload, the webhook should still fire.
func TestHandleWebhook_EmptyBody(t *testing.T) {
	mock := &mockProducer{}
	router := newTestRouter(mock)

	req := httptest.NewRequest(http.MethodPost, "/hooks/relay-1", bytes.NewReader([]byte{}))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if mock.callCount != 1 {
		t.Errorf("expected 1 publish call, got %d", mock.callCount)
	}
}
