package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSecretNotFound = errors.New("secret not found")

type SecretStore struct {
	db *pgxpool.Pool
}

func NewSecretStore(db *pgxpool.Pool) *SecretStore {
	return &SecretStore{db: db}
}

func (s *SecretStore) Create(ctx context.Context, req models.CreateSecretRequest) (*models.Secret, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("secret name if required")
	}
	if strings.TrimSpace(req.Value) == "" {
		return nil, fmt.Errorf("secret value is required")
	}

	id := uuid.NewString()
	now := time.Now()
	query := `INSERT INTO secrets (id, user_id, name, value, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, user_id, name, created_at, updated_at`

	var secret models.Secret
	err := s.db.QueryRow(ctx, query, id, req.UserID, req.Name, req.Value, now, now).Scan(
		&secret.ID,
		&secret.UserID,
		&secret.Name,
		&secret.CreatedAt,
		&secret.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("secret with name %q already exists for this user", req.Name)
		}
		return nil, fmt.Errorf("insert secret: %w", err)
	}
	return &secret, nil
}

func (s *SecretStore) ListByUser(ctx context.Context, userID string) ([]models.Secret, error) {
	query := `SELECT id, user_id, name, created-at, updated-at FROM secrets WHERE user_id = $1 ORDER BY name ASC`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query secrets: %w", err)
	}
	defer rows.Close()

	secrets := make([]models.Secret, 0)
	for rows.Next() {
		var sec models.Secret
		if err := rows.Scan(&sec.ID, &sec.UserID, &sec.Name, &sec.CreatedAt, &sec.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan secret: %w", err)
		}
		secrets = append(secrets, sec)
	}
	return secrets, nil
}

func (s *SecretStore) Delete(ctx context.Context, usreID, secretID string) error {
	query := `DELETE FROM secrets WHERE id=$1 AND user_id=$2`
	result, err := s.db.Exec(ctx, query, secretID, usreID)
	if err != nil {
		return fmt.Errorf("delete secret: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrSecretNotFound
	}
	return nil
}
