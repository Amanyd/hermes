package actions

import (
	"testing"
)

// Checks that every registered type is recognized and unknown types are rejected
func TestIsValid(t *testing.T) {
	validTypes := []string{"debug_log", "discord_send", "slack_send", "email_send", "http_request"}
	for _, actionType := range validTypes {
		if !IsValidType(actionType) {
			t.Errorf("IsValidType(%q) = false, want true", actionType)
		}
	}
	invalidTypes := []string{"android_send", "sms", "unknown"}
	for _, actionType := range invalidTypes {
		if IsValidType(actionType) {
			t.Errorf("IsValidType(%q) = true, want false", actionType)
		}
	}
}

// Checks that the Types() helper returns a sorted list of all registered action types
func TestTypes(t *testing.T) {
	types := Types()
	if len(types) != 5 {
		t.Fatalf("expected 5 types, got %d: %v", len(types), types)
	}
	for i := 1; i < len(types); i++ {
		if types[i] < types[i-1] {
			t.Errorf("types not sorted: %v", types)
			break
		}
	}
}

// Verifies the debug_log has no required fields so even empty config passes validation
func TestValidationConfig_DebugLog(t *testing.T) {
	err := ValidateConfig("debug_log", map[string]any{})
	if err != nil {
		t.Errorf("debug_log with empty config should pass, got %v", err)
	}
	err = ValidateConfig("debug_log", map[string]any{"prefix": "TEST"})
	if err != nil {
		t.Errorf("debug_log with prefix should pass, got %v", err)
	}
}

// Verifies ValidateConfig works as intended for discord_send
func TestValidateConfig_DiscordSend(t *testing.T) {
	err := ValidateConfig("discord_send", map[string]any{})
	if err == nil {
		t.Error("discord_send with empty config should fail")
	}
	err = ValidateConfig("discord_send", map[string]any{"webhook_url": "https://discord.com/api/webhook/1234/abc"})
	if err != nil {
		t.Errorf("discord_send with webhook_url should pass, got %v", err)
	}
	err = ValidateConfig("discord_send", map[string]any{"webhook_url_ref": "https://discord.com/api/webhook/1234/abc"})
	if err != nil {
		t.Errorf("discord_send with webhook_url_url should pass, got %v", err)
	}
	err = ValidateConfig("discord_send", map[string]any{"webhug": "masti"})
	if err == nil {
		t.Error("discord_send with invalid fields passed")
	}
}

// Verifies ValidateConfig works as intended for slack_send
func TestValidateConfig_SlackSend(t *testing.T) {
	err := ValidateConfig("slack_send", map[string]any{})
	if err == nil {
		t.Error("slack_send with empty config should fail")
	}
	err = ValidateConfig("slack_send", map[string]any{"webhook_url": "https://slack.com/api/webhook/1234/abc"})
	if err != nil {
		t.Errorf("slack_send with webhook_url should pass, got %v", err)
	}
	err = ValidateConfig("slack_send", map[string]any{"webhook_url_ref": "https://slack.com/api/webhook/1234/abc"})
	if err != nil {
		t.Errorf("slack_send with webhook_url_url should pass, got %v", err)
	}
	err = ValidateConfig("slack_send", map[string]any{"webhug": "masti"})
	if err == nil {
		t.Error("slack_send with invalid fields passed")
	}
}

// Checks the http_request schema including the Extra validation for headers and method
func TestValidateConfig_HTTPRequest(t *testing.T) {
	err := ValidateConfig("http_request", map[string]any{})
	if err == nil {
		t.Error("http_request with empty config should fail")
	}
	err = ValidateConfig("http_request", map[string]any{
		"url": "https://fmhy.net",
	})
	if err != nil {
		t.Errorf("http_request with url should pass, got %v", err)
	}
	err = ValidateConfig("http_request", map[string]any{
		"url_ref": "https://open-slum.com",
	})
	if err != nil {
		t.Errorf("http_request with url_ref should pass, got %v", err)
	}

	err = ValidateConfig("http_request", map[string]any{
		"url":    "https://ye.com",
		"method": "INVALID",
	})
	if err == nil {
		t.Error("http_request with invalid method should fail")
	}

	// headers must be a JSON object, not a string
	err = ValidateConfig("http_request", map[string]any{
		"url":     "https://ye.com",
		"headers": "string-header",
	})
	if err == nil {
		t.Error("http_request with string headers should fail")
	}

	err = ValidateConfig("http_request", map[string]any{
		"url":     "https://songspk.com",
		"method":  "POST",
		"headers": map[string]any{"X-Custom": "value"},
	})
	if err != nil {
		t.Errorf("http_request with valid headers should pass, got: %v", err)
	}
}

// Checks all required fields for the email integration
func TestValidateConfig_EmailSend(t *testing.T) {
	err := ValidateConfig("email_send", map[string]any{})
	if err == nil {
		t.Error("email_send with empty config should fail")
	}

	err = ValidateConfig("email_send", map[string]any{
		"api_key": "re_1234",
		"from":    "noreply@hermes.dev",
	})
	if err == nil {
		t.Error("email_send missing 'to' should fail")
	}

	err = ValidateConfig("email_send", map[string]any{
		"api_key": "1234abc",
		"from":    "email@email.com",
		"to":      "reply@reply.com",
	})
	if err != nil {
		t.Errorf("email_send with all fields should pass, got: %v", err)
	}
}

// Unregistered action type should fail and return an error
func TestValidateConfig_UnknownType(t *testing.T) {
	err := ValidateConfig("nonexistent", map[string]any{})
	if err == nil {
		t.Error("expected error for unknown action type")
	}
}
