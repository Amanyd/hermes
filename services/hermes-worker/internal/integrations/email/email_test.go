package email

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/oauth"
)

// Satisfies ConnectionResolver for testing.
type fakeConnStore struct {
	provider     string
	accessToken  string
	refreshToken string
	email        string
	expiry       time.Time
	getErr       error

	// Track UpdateConnectionTokens calls
	updatedID           string
	updatedAccessToken  string
	updatedRefreshToken string
	updatedExpiry       time.Time
	updateErr           error
}

func (f *fakeConnStore) GetConnection(_ context.Context, connectionID string) (string, string, string, string, time.Time, error) {
	if f.getErr != nil {
		return "", "", "", "", time.Time{}, f.getErr
	}
	return f.provider, f.accessToken, f.refreshToken, f.email, f.expiry, nil
}

func (f *fakeConnStore) UpdateConnectionTokens(_ context.Context, id, access, refresh string, expiry time.Time) error {
	f.updatedID = id
	f.updatedAccessToken = access
	f.updatedRefreshToken = refresh
	f.updatedExpiry = expiry
	return f.updateErr
}

// Satisfies oauth.Provider for testing.
type fakeProvider struct {
	sendErr    error
	refreshErr error

	// Track calls
	sendCalled    bool
	refreshCalled bool
	lastFrom      string
	lastTo        string
	lastSubject   string
	lastBody      string

	// What Refresh returns
	refreshResult *oauth.Tokens
}

func (f *fakeProvider) AuthURL(_ string) string { return "" }

func (f *fakeProvider) Exchange(_ context.Context, _ string) (*oauth.Tokens, error) {
	return nil, nil
}

func (f *fakeProvider) Refresh(_ context.Context, _ string) (*oauth.Tokens, error) {
	f.refreshCalled = true
	if f.refreshErr != nil {
		return nil, f.refreshErr
	}
	return f.refreshResult, nil
}

func (f *fakeProvider) SendEmail(_ context.Context, _, from, to, subject, body string) error {
	f.sendCalled = true
	f.lastFrom = from
	f.lastTo = to
	f.lastSubject = subject
	f.lastBody = body
	return f.sendErr
}

// Happy path test
func TestExecute_Success(t *testing.T) {
	fp := &fakeProvider{}
	fc := &fakeConnStore{
		provider:     "google",
		accessToken:  "valid-token",
		refreshToken: "refresh-token",
		email:        "greetu@gmail.com",
		expiry:       time.Now().Add(1 * time.Hour),
	}

	sender := New(map[string]oauth.Provider{"google": fp}, fc)

	cfg := map[string]any{
		"connection_id": "conn-1",
		"to":            "amaanu@example.com",
		"subject":       "Hello",
		"body":          "World",
	}

	_, err := sender.Execute(context.Background(), cfg, []byte(`{}`), nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !fp.sendCalled {
		t.Error("expected SendEmail to be called")
	}
	if fp.lastFrom != "greetu@gmail.com" {
		t.Errorf("expected from='alice@gmail.com', got %q", fp.lastFrom)
	}
	if fp.lastTo != "amaanu@example.com" {
		t.Errorf("expected to='bob@example.com', got %q", fp.lastTo)
	}
	if fp.lastSubject != "Hello" {
		t.Errorf("expected subject='Hello', got %q", fp.lastSubject)
	}
}

// Verifies validation.
func TestExecute_MissingConnectionID(t *testing.T) {
	sender := New(nil, &fakeConnStore{})
	_, err := sender.Execute(context.Background(), map[string]any{"to": "a@g.com"}, nil, nil)
	if err == nil {
		t.Error("expected error for missing connection_id")
	}
}

// Verifies validation.
func TestExecute_MissingTo(t *testing.T) {
	sender := New(nil, &fakeConnStore{})
	_, err := sender.Execute(context.Background(), map[string]any{"connection_id": "c1"}, nil, nil)
	if err == nil {
		t.Error("expected error for missing to")
	}
}

// Verifies store error propagation.
func TestExecute_ConnectionNotFound(t *testing.T) {
	fc := &fakeConnStore{getErr: fmt.Errorf("connection not found: conn-999")}
	sender := New(map[string]oauth.Provider{}, fc)

	_, err := sender.Execute(context.Background(), map[string]any{
		"connection_id": "conn-999",
		"to":            "a@b.com",
	}, nil, nil)
	if err == nil {
		t.Error("expected error for missing connection")
	}
}

// Verifies that an expired token triggers a refresh before sending.
func TestExecute_TokenRefresh(t *testing.T) {
	fp := &fakeProvider{
		refreshResult: &oauth.Tokens{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			Expiry:       time.Now().Add(1 * time.Hour),
		},
	}
	fc := &fakeConnStore{
		provider:     "google",
		accessToken:  "expired-token",
		refreshToken: "old-refresh",
		email:        "alice@gmail.com",
		expiry:       time.Now().Add(-10 * time.Minute),
	}

	sender := New(map[string]oauth.Provider{"google": fp}, fc)

	_, err := sender.Execute(context.Background(), map[string]any{
		"connection_id": "conn-1",
		"to":            "bob@example.com",
	}, []byte(`{}`), nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !fp.refreshCalled {
		t.Error("expected Refresh to be called for expired token")
	}

	if fc.updatedAccessToken != "new-access-token" {
		t.Errorf("expected updated access token='new-access-token', got %q", fc.updatedAccessToken)
	}
}

// Verifies that a refresh error is propagated.
func TestExecute_RefreshFailure(t *testing.T) {
	fp := &fakeProvider{
		refreshErr: fmt.Errorf("refresh token revoked"),
	}
	fc := &fakeConnStore{
		provider: "google",
		expiry:   time.Now().Add(-10 * time.Minute), // expired
	}

	sender := New(map[string]oauth.Provider{"google": fp}, fc)

	_, err := sender.Execute(context.Background(), map[string]any{
		"connection_id": "conn-1",
		"to":            "a@b.com",
	}, nil, nil)
	if err == nil {
		t.Error("expected error when refresh fails")
	}
}

// Verifies error when provider isn't registered.
func TestExecute_UnsupportedProvider(t *testing.T) {
	fc := &fakeConnStore{
		provider: "badabingbadaboom",
		expiry:   time.Now().Add(1 * time.Hour),
	}

	sender := New(map[string]oauth.Provider{"google": &fakeProvider{}}, fc)

	_, err := sender.Execute(context.Background(), map[string]any{
		"connection_id": "conn-1",
		"to":            "a@b.com",
	}, nil, nil)
	if err == nil {
		t.Error("expected error for unsupported provider")
	}
}
