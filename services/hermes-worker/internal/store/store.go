package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/encryptor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RelayAction struct {
	OrderIndex int
	ActionType string
	Config     map[string]any
}

type Store struct {
	db        *pgxpool.Pool
	encryptor *encryptor.Encryptor
}

var (
	ErrRelayNotFound      = errors.New("relay not found")
	ErrNoActions          = errors.New("no actions configured for relay")
	ErrSecretNotFound     = errors.New("secret not found")
	ErrConnectionNotFound = errors.New("connection not found")
)

func NewStore(dbURL string, enc *encryptor.Encryptor) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to db: %w", err)
	}
	return &Store{db: pool, encryptor: enc}, nil
}

func (s *Store) GetRelayActions(ctx context.Context, relayID string) ([]RelayAction, error) {
	query := `SELECT a.action_type, a.config, a.order_index
	FROM relays r
	JOIN relay_actions a ON r.id=a.relay_id
	WHERE r.id=$1 AND r.is_active=true
	ORDER BY a.order_index ASC`

	rows, err := s.db.Query(ctx, query, relayID)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}
	defer rows.Close()

	actions := make([]RelayAction, 0)
	for rows.Next() {
		var act RelayAction
		var configBytes []byte
		if err := rows.Scan(&act.ActionType, &configBytes, &act.OrderIndex); err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}
		if err := json.Unmarshal(configBytes, &act.Config); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
		actions = append(actions, act)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	if len(actions) == 0 {
		return nil, ErrNoActions
	}
	return actions, nil
}

func (s *Store) GetRelayOwner(ctx context.Context, relayID string) (string, error) {
	var userID string
	err := s.db.QueryRow(ctx, `SELECT user_id FROM relays WHERE id = $1`, relayID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrRelayNotFound
	}
	if err != nil {
		return "", fmt.Errorf("query relay owner: %w", err)
	}
	return userID, nil
}

func (s *Store) RegisterEvent(ctx context.Context, relayID, eventID string) (bool, error) {
	if eventID == "" {
		return true, nil
	}
	query := `INSERT INTO processed_events (relay_id, event_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
	tag, err := s.db.Exec(ctx, query, relayID, eventID)
	if err != nil {
		return false, fmt.Errorf("dedupe insert failed: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) LogExecution(ctx context.Context, relayID string, eventID string, status string, details string, payload []byte) error {
	query := `INSERT INTO execution_logs(relay_id, event_id, status,payload,error_message,executed_at)
	VALUES($1,$2,$3,$4,$5,NOW())`

	var payloadJSON any
	if len(payload) > 0 {
		payloadJSON = json.RawMessage(payload)
	}

	var errorMessage any
	if status != "success" && details != "" {
		errorMessage = details
	}

	_, err := s.db.Exec(ctx, query, relayID, eventID, status, payloadJSON, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to write execution log: %w", err)
	}
	return nil
}

func (s *Store) ResolveSecret(ctx context.Context, userID, secretName string) (string, error) {
	var encrypted string
	err := s.db.QueryRow(ctx,
		`SELECT value FROM secrets WHERE user_id = $1 AND name = $2`,
		userID, secretName,
	).Scan(&encrypted)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("%w: %s", ErrSecretNotFound, secretName)
	}
	if err != nil {
		return "", fmt.Errorf("resolve secret: %w", err)
	}
	plaintext, err := s.encryptor.Decrypt(encrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt secret %q: %w", secretName, err)
	}
	return plaintext, nil
}

func (s *Store) GetConnection(ctx context.Context, connectionID string) (provider, accessToken, refreshToken, accountEmail string, expiry time.Time, err error) {
	var encAccess, encRefresh string
	err = s.db.QueryRow(ctx, `SELECT provider, access_token, refresh_token, account_email, token_expiry FROM connections WHERE id = $1`,
		connectionID).Scan(&provider, &encAccess, &encRefresh, &accountEmail, &expiry)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", "", time.Time{}, fmt.Errorf("%w: %s", ErrConnectionNotFound, connectionID)
	}
	if err != nil {
		return "", "", "", "", time.Time{}, fmt.Errorf("query connection: %w", err)
	}

	accessToken, err = s.encryptor.Decrypt(encAccess)
	if err != nil {
		return "", "", "", "", time.Time{}, fmt.Errorf("decrypt access token: %w", err)
	}
	refreshToken, err = s.encryptor.Decrypt(encRefresh)
	if err != nil {
		return "", "", "", "", time.Time{}, fmt.Errorf("decrypt refresh token: %w", err)
	}

	return provider, accessToken, refreshToken, accountEmail, expiry, nil
}

func (s *Store) UpdateConnectionTokens(ctx context.Context, connectionID, accessToken, refreshToken string, expiry time.Time) error {
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
