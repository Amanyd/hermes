package email

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/oauth"
)

type ConnectionResolver interface {
	GetConnection(ctx context.Context, connectionID string) (provider, accessToken, refreshToken, accountEmail string, expiry time.Time, err error)
	UpdateConnectionTokens(ctx context.Context, connectionID, accessToken, refreshToken string, expiry time.Time) error
}

type Sender struct {
	providers map[string]oauth.Provider
	connStore ConnectionResolver
}

func New(providers map[string]oauth.Provider, connStore ConnectionResolver) *Sender {
	return &Sender{
		providers: providers,
		connStore: connStore,
	}
}

func (s *Sender) Execute(ctx context.Context, cfg map[string]any, payload []byte) error {
	connectionID, _ := cfg["connection_id"].(string)
	if connectionID == "" {
		return fmt.Errorf("email_send: missing connection_id")
	}
	to, _ := cfg["to"].(string)
	if to == "" {
		return fmt.Errorf("email_send: missing to")
	}
	subject, _ := cfg["subject"].(string)
	if subject == "" {
		subject = "Hermes Notification"
	}

	body, _ := cfg["body"].(string)
	if body == "" {
		body = fmt.Sprintf("Payload received:\n%s", string(payload))
	}

	providerName, accessToken, refreshToken, accountEmail, expiry, err := s.connStore.GetConnection(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("email_send: load connection: %w", err)
	}
	provider, ok := s.providers[providerName]
	if !ok {
		return fmt.Errorf("email_send: unsupported provider %q", providerName)
	}

	if time.Now().After(expiry.Add(-1 * time.Minute)) {
		refreshed, refreshErr := provider.Refresh(ctx, refreshToken)
		if refreshErr != nil {
			return fmt.Errorf("email_send: token refresh failed: %w", refreshErr)
		}
		accessToken = refreshed.AccessToken
		if updateErr := s.connStore.UpdateConnectionTokens(ctx, connectionID, refreshed.AccessToken, refreshed.RefreshToken, refreshed.Expiry); updateErr != nil {
			_ = updateErr
		}
	}

	recipients := strings.SplitSeq(to, "")
	for recipient := range recipients {
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			continue
		}
		if err := provider.SendEmail(ctx, accessToken, accountEmail, recipient, subject, body); err != nil {
			return fmt.Errorf("email_send to %q:%w", recipient, err)
		}
	}
	return nil
}
