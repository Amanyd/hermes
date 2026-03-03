package api

import (
	"context"
	"log/slog"

	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
)

type RelayStorer interface {
	CreateRelay(ctx context.Context, req models.CreateRelayRequest) (*models.RelayWithActions, error)
	GetAllRelays(ctx context.Context, userID string) ([]models.Relay, error)
	GetRelay(ctx context.Context, relayID, userID string) (*models.RelayWithActions, error)
	UpdateRelay(ctx context.Context, relayID, userID string, req models.UpdateRelayRequest) (*models.Relay, error)
	DeleteRelay(ctx context.Context, relayID, userID string) error
	GetLogs(ctx context.Context, relayID, userID string, limit int) ([]models.ExecutionLog, error)
}

type SecretStorer interface {
	Create(ctx context.Context, req models.CreateSecretRequest) (*models.Secret, error)
	ListByUser(ctx context.Context, userID string) ([]models.Secret, error)
	Delete(ctx context.Context, userID, secretID string) error
}

type UserStorer interface {
	CreateUser(ctx context.Context, username, email, passwordHash string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type Handler struct {
	store       RelayStorer
	secretStore SecretStorer
	userStore   UserStorer
	logger      *slog.Logger
	baseURL     string
	jwtSecret   string
}

func NewHandler(s RelayStorer, ss SecretStorer, us UserStorer, jwtSecret string, logger *slog.Logger) *Handler {
	return &Handler{store: s, secretStore: ss, userStore: us, jwtSecret: jwtSecret, logger: logger, baseURL: "http://localhost:8080"}
}
