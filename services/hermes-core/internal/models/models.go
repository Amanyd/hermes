package models

import "time"

type TriggerType string

const (
	TriggerWebhook TriggerType = "webhook"
	TriggerManual  TriggerType = "manual"
	TriggerCron    TriggerType = "cron"
)

type CreateRelayRequest struct {
	Name          string                   `json:"name"`
	UserID        string                   `json:"user_id"`
	Description   string                   `json:"description"`
	TriggerType   TriggerType              `json:"trigger_type,omitempty"`
	TriggerConfig map[string]any           `json:"trigger_config,omitempty"`
	Actions       []CreateRelayActionInput `json:"actions"`
}

type CreateRelayActionInput struct {
	ActionType string         `json:"action_type"`
	Config     map[string]any `json:"config"`
	OrderIndex int            `json:"order_index"`
}

type UpdateRelayRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type UpdateRelayActionsRequest struct {
	Actions []CreateRelayActionInput `json:"actions"`
}

type Execution struct {
	ID             string          `json:"id"`
	RelayID        string          `json:"relay_id"`
	EventID        string          `json:"event_id,omitempty"`
	Status         string          `json:"status"`
	TriggerPayload map[string]any  `json:"trigger_payload,omitempty"`
	ErrorMessage   string          `json:"error_message,omitempty"`
	StartedAt      time.Time       `json:"started_at"`
	FinishedAt     *time.Time      `json:"finished_at,omitempty"`
	Steps          []ExecutionStep `json:"steps,omitempty"`
}

type ExecutionStep struct {
	ID           string         `json:"id"`
	ExecutionID  string         `json:"execution_id"`
	OrderIndex   int            `json:"order_index"`
	ActionType   string         `json:"action_type"`
	Status       string         `json:"status"`
	Input        map[string]any `json:"input,omitempty"`
	Output       map[string]any `json:"output,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
	StartedAt    time.Time      `json:"started_at"`
	FinishedAt   *time.Time     `json:"finished_at,omitempty"`
}

type Relay struct {
	ID            string         `json:"id"`
	UserID        string         `json:"user_id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	WebhookPath   string         `json:"webhook_path"`
	WebhookURL    string         `json:"webhook_url"`
	TriggerType   TriggerType    `json:"trigger_type"`
	TriggerConfig map[string]any `json:"trigger_config,omitempty"`
	IsActive      bool           `json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type RelayWithActions struct {
	Relay
	Actions []RelayAction `json:"actions"`
}

type RelayAction struct {
	ID         string         `json:"id"`
	RelayID    string         `json:"relay_id"`
	ActionType string         `json:"action_type"`
	Config     map[string]any `json:"config"`
	OrderIndex int            `json:"order_index"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type Secret struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateSecretRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// Auth models

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// OAuth models

type Connection struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Provider     string    `json:"provider"`
	AccountEmail string    `json:"account_email"`
	Scopes       string    `json:"scopes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ConnectionInternal struct {
	Connection
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	TokenExpiry  time.Time `json:"-"`
}

type OAuthCallbackParams struct {
	Provider string
	Code     string
	State    string
}
