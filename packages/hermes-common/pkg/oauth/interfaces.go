package oauth

import (
	"context"
	"time"
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
	Email        string
}

type Provider interface {
	AuthURL(state string) string
	Exchange(ctx context.Context, code string) (*Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (*Tokens, error)
	SendEmail(ctx context.Context, accessToken string, from, to, subject, body string) error
}

const (
	ProviderGoogle    = "google"
	ProviderMicrosoft = "microsoft"
)

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}
