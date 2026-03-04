package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/encryptor"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrConnectionNotFound = errors.New("connection not found")

type ConnectionStore struct {
	db        *pgxpool.Pool
	encryptor *encryptor.Encryptor
}

func NewConnectionStore(db *pgxpool.Pool, enc *encryptor.Encryptor) *ConnectionStore {
	return &ConnectionStore{db: db, encryptor: enc}
}

// Creates or Updates a connection after a successful OAuth callback
// If the provider+email combo exists, it updates the tokens
func (s *ConnectionStore) Upsert(ctx context.Context, userID, provider, accountEmail, accessToken, refreshToken, scopes string,
	expiry time.Time) (*models.Connection, error) {
	encAccess, err := s.encryptor.Encrypt(accessToken)
	if err != nil {
		return nil, fmt.Errorf("encrypt refresh token: %w", err)
	}
	encRefresh, err := s.encryptor.Encrypt(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("encrypt refresh token: %w", err)
	}
	id := uuid.NewString()
	now := time.Now()

	query := `
			INSERT INTO connections (id, user_id, provider, account_email, access_token, refresh_token, token_expiry, scopes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (user_id, provider, account_email)
			DO UPDATE SET access_token = $5, refresh_token = $6, token_expiry = $7, scopes = $8, updated_at = $10
			RETURNING id, user_id, provider, account_email, scopes, created_at, updated_at`
	var conn models.Connection
	err = s.db.QueryRow(ctx, query, id, userID, provider, accountEmail, encAccess, encRefresh, expiry, scopes, now, now).Scan(
		&conn.ID, &conn.UserID, &conn.Provider, &conn.AccountEmail, &conn.Scopes, &conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert connection: %w", err)
	}
	return &conn, nil
}

// Returns all connections for a user
func (s *ConnectionStore) ListByUser(ctx context.Context, userID string) ([]models.Connection, error) {
	query := `SELECT id, user_id, provider, account_email, scopes, created_at, updated_at
		FROM connections WHERE user_id = $1 ORDER BY provider, account_email`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query connections: %w", err)
	}
	defer rows.Close()

	conns := make([]models.Connection, 0)
	for rows.Next() {
		var c models.Connection
		if err := rows.Scan(&c.ID, &c.UserID, &c.Provider, &c.AccountEmail, &c.Scopes, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan connection: %w", err)
		}
		conns = append(conns, c)
	}
	return conns, nil
}

// Removes a connection.
func (s *ConnectionStore) Delete(ctx context.Context, userID, connectionID string) error {
	result, err := s.db.Exec(ctx, `DELETE FROM connections WHERE id = $1 AND user_id = $2`, connectionID, userID)
	if err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrConnectionNotFound
	}
	return nil
}

// Returns the full connection including encrypted tokens.
// Used by the Worker only, not the API.
func (s *ConnectionStore) GetInternal(ctx context.Context, connectionID string) (*models.ConnectionInternal, error) {
	query := `SELECT id, user_id, provider, account_email, access_token, refresh_token, token_expiry, scopes, created_at, updated_at
		FROM connections WHERE id = $1`

	var c models.ConnectionInternal
	var encAccess, encRefresh string
	err := s.db.QueryRow(ctx, query, connectionID).Scan(
		&c.ID, &c.UserID, &c.Provider, &c.AccountEmail, &encAccess, &encRefresh, &c.TokenExpiry, &c.Scopes, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConnectionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query connection: %w", err)
	}

	c.AccessToken, err = s.encryptor.Decrypt(encAccess)
	if err != nil {
		return nil, fmt.Errorf("decrypt access token: %w", err)
	}
	c.RefreshToken, err = s.encryptor.Decrypt(encRefresh)
	if err != nil {
		return nil, fmt.Errorf("decrypt refresh token: %w", err)
	}

	return &c, nil
}

// Updates the tokens after a refresh (called by the worker).
func (s *ConnectionStore) UpdateTokens(ctx context.Context, connectionID, accessToken, refreshToken string, expiry time.Time) error {
	encAccess, err := s.encryptor.Encrypt(accessToken)
	if err != nil {
		return fmt.Errorf("encrypt access token: %w", err)
	}
	encRefresh, err := s.encryptor.Encrypt(refreshToken)
	if err != nil {
		return fmt.Errorf("encrypt refresh token: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`UPDATE connections SET access_token = $1, refresh_token = $2, token_expiry = $3, updated_at = NOW() WHERE id = $4`,
		encAccess, encRefresh, expiry, connectionID)
	return err
}
