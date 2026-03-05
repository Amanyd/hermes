package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type MicrosoftProvider struct {
	cfg    ProviderConfig
	client *http.Client
}

func NewMicrosoftProvider(cfg ProviderConfig) *MicrosoftProvider {
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{
			"https://graph.microsoft.com/Mail.Send",
			"https://graph.microsoft.com/User.Read",
			"offline_access",
		}
	}
	return &MicrosoftProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *MicrosoftProvider) AuthURL(state string) string {
	params := url.Values{
		"client_id":     {m.cfg.ClientID},
		"redirect_uri":  {m.cfg.RedirectURL},
		"response_type": {"code"},
		"scope":         {strings.Join(m.cfg.Scopes, " ")},
		"state":         {state},
	}
	return "https://login.microsoftonline.com/common/oauth2/v2.0/authorize?" + params.Encode()
}

func (m *MicrosoftProvider) Exchange(ctx context.Context, code string) (*Tokens, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {m.cfg.ClientID},
		"client_secret": {m.cfg.ClientSecret},
		"redirect_uri":  {m.cfg.RedirectURL},
		"grant_type":    {"authorization_code"},
	}

	resp, err := m.client.PostForm("https://login.microsoftonline.com/common/oauth2/v2.0/token", data)
	if err != nil {
		return nil, fmt.Errorf("microsoft token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("microsoft token exchange failed (%d): %s", resp.StatusCode, body)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode microsoft token response: %w", err)
	}

	email, err := m.fetchEmail(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("fetch microsoft email: %w", err)
	}

	return &Tokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Email:        email,
	}, nil
}

func (m *MicrosoftProvider) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	data := url.Values{
		"refresh_token": {refreshToken},
		"client_id":     {m.cfg.ClientID},
		"client_secret": {m.cfg.ClientSecret},
		"grant_type":    {"refresh_token"},
		"scope":         {strings.Join(m.cfg.Scopes, " ")},
	}

	resp, err := m.client.PostForm("https://login.microsoftonline.com/common/oauth2/v2.0/token", data)
	if err != nil {
		return nil, fmt.Errorf("microsoft token refresh: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode microsoft refresh response: %w", err)
	}

	return &Tokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

func (m *MicrosoftProvider) SendEmail(ctx context.Context, accessToken string, from, to, subject, body string) error {
	payload := map[string]any{
		"message": map[string]any{
			"subject": subject,
			"body": map[string]any{
				"contentType": "Text",
				"content":     body,
			},
			"toRecipients": []map[string]any{
				{"emailAddress": map[string]string{"address": to}},
			},
		},
	}
	payloadJSON, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://graph.microsoft.com/v1.0/me/sendMail",
		bytes.NewReader(payloadJSON))
	if err != nil {
		return fmt.Errorf("build outlook request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("outlook send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("graph API returned %d: %s", resp.StatusCode, respBody)
	}

	return nil
}

func (m *MicrosoftProvider) fetchEmail(ctx context.Context, accessToken string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.microsoft.com/v1.0/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var profile struct {
		Mail string `json:"mail"`
		UPN  string `json:"userPrincipalName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return "", err
	}
	if profile.Mail != "" {
		return profile.Mail, nil
	}
	return profile.UPN, nil
}
