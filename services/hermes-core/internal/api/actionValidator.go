package api

import (
	"fmt"
	"strings"
)

var allowedActionType = map[string]bool{
	"debug_log":    true,
	"discord_send": true,
	"email_send":   true,
	"slack_send":   true,
	"http_request": true,
}

func IsValidActionType(actionType string) bool {
	return allowedActionType[actionType]
}

func ValidateActionConfig(actionType string, config map[string]any) error {
	switch actionType {
	case "debug_log":
		return validateDebugLog(config)
	case "discord_send":
		return validateDiscordSend(config)
	case "slack_send":
		return validateSlackSend(config)
	case "email_send":
		return validateEmailSend(config)
	case "http_request":
		return validateHTTPRequest(config)
	default:
		return fmt.Errorf("unknown action type: %s", actionType)
	}
}

func validateDebugLog(_ map[string]any) error {
	return nil
}

func validateDiscordSend(cfg map[string]any) error {
	if !hasOneOf(cfg, "webhook_url", "webhook_url_ref") {
		return fmt.Errorf("discord_send requires webhook_url or wehook_url_ref")
	}
	return nil
}

func validateSlackSend(cfg map[string]any) error {
	if !hasOneOf(cfg, "webhook_url", "webhook_url_ref") {
		return fmt.Errorf("slack_send requires webhook_url or wehook_url_ref")
	}
	return nil
}

func validateHTTPRequest(cfg map[string]any) error {
	if !hasOneOf(cfg, "url", "url_ref") {
		return fmt.Errorf("http_request requires url or url_ref")
	}
	if method, ok := cfg["method"].(string); ok {
		allowed := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
		}
		if !allowed[strings.ToUpper(method)] {
			return fmt.Errorf("http_request has no valid method")
		}
	}
	if headers, ok := cfg["headers"]; ok {
		if _, ok := headers.(map[string]any); !ok {
			return fmt.Errorf("http_request headers must be an object")
		}
	}
	return nil
}

func validateEmailSend(cfg map[string]any) error {
	if !hasOneOf(cfg, "api_key", "api_key_ref") {
		return fmt.Errorf("email send requires api_key or api_key_ref")
	}
	if !hasString(cfg, "from") {
		return fmt.Errorf("email_send requires from")
	}
	if !hasString(cfg, "to") {
		return fmt.Errorf("email-send requires to")
	}
	return nil
}

func hasOneOf(cfg map[string]any, keys ...string) bool {
	for _, key := range keys {
		if v, ok := cfg[key].(string); ok && strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

func hasString(cfg map[string]any, key string) bool {
	v, ok := cfg[key].(string)
	return ok && strings.TrimSpace(v) != ""
}
