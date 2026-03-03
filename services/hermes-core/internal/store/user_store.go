package store

import (
	"context"
	"errors"
	"fmt"
	"time"
	"strings"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	db *pgxpool.Pool
}

var ErrUserNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already taken")
var ErrUsernameTaken = errors.New("username already taken")

func NewUserStore(db *pgxpool.Pool) *UserStore {
	return &UserStore{db: db}
}

func (s * UserStore) CreateUser(ctx context.Context, username, email, passwordHash string) (*models.User, error){
	queryUser := `INSERT INTO users (id, username, email, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, username, email, created_at, updated_at`
	var user models.User
	userID := uuid.New().String()

	err := s.db.QueryRow(ctx, queryUser, userID, username, email, passwordHash, time.Now(), time.Now()).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "users_username_key") {
			return nil, ErrUsernameTaken
		}
		if strings.Contains(err.Error(), "users_email_key") {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}
	return &user, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = $1`
	var user models.User
	err := s.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}
	return &user, nil
}