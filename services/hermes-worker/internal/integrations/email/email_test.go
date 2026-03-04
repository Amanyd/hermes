package email

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Verifies the happy path for email send
func TestExecute_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer re_test_key" {
			t.Errorf("expected 'Bearer re_test_key', got %q", auth)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	sender := New()
	sender.client = srv.Client()

	cfg := map[string]any{
		"api_key": "re_test_key",
		"from":    "geetuuuu@gmail.com",
		"to":      "meinhu@gmail.com",
		"subject": "Test Subj",
		"body":    "Hello World",
	}
}
