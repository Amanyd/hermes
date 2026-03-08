package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
			slog.String("path", r.URL.Path),
		)
		h.respondError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
		return
	}

	if strings.TrimSpace(req.Username) == "" {
		h.respondError(w, http.StatusBadRequest, "Username is required", "VALIDATION_ERROR")
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		h.respondError(w, http.StatusBadRequest, "Email is required", "VALIDATION_ERROR")
		return
	}
	if len(req.Password) < 8 {
		h.respondError(w, http.StatusBadRequest, "Password must be at least 8 characters", "VALIDATION_ERROR")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("failed to hash password", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to register user", "INTERNAL_ERROR")
		return
	}

	user, err := h.userStore.CreateUser(r.Context(), req.Username, req.Email, string(hashedPassword))
	if err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			h.respondError(w, http.StatusConflict, "Email already taken", "EMAIL_TAKEN")
			return
		}
		if errors.Is(err, store.ErrUsernameTaken) {
			h.respondError(w, http.StatusConflict, "Username already taken", "USERNAME_TAKEN")
			return
		}
		h.logger.Error("failed to create user", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to register user", "DB_ERROR")
		return
	}

	h.logger.Info("user registered",
		slog.String("user_id", user.ID),
		slog.String("username", user.Username),
	)
	signedToken, err := h.issueToken(user.ID)
	if err != nil {
		h.logger.Error("failed to generate JWT token", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to login", "TOKEN_ERROR")
		return
	}

	h.respondSuccess(w, http.StatusCreated, "User registered successfully", models.AuthResponse{
		Token: signedToken,
		User:  *user,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
			slog.String("path", r.URL.Path),
		)
		h.respondError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		h.respondError(w, http.StatusBadRequest, "Email is required", "VALIDATION_ERROR")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		h.respondError(w, http.StatusBadRequest, "Password is required", "VALIDATION_ERROR")
		return
	}

	user, err := h.userStore.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			h.respondError(w, http.StatusUnauthorized, "Invalid email or password", "AUTH_FAILED")
			return
		}
		h.logger.Error("failed to fetch user by email", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to login", "DB_ERROR")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.respondError(w, http.StatusUnauthorized, "Invalid email or password", "AUTH_FAILED")
		return
	}

	signedToken, err := h.issueToken(user.ID)
	if err != nil {
		h.logger.Error("failed to generate JWT token", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to login", "TOKEN_ERROR")
		return
	}

	h.logger.Info("user logged in",
		slog.String("user_id", user.ID),
		slog.String("username", user.Username),
	)
	h.respondSuccess(w, http.StatusOK, "Login successful", models.AuthResponse{
		Token: signedToken,
		User:  *user,
	})
}

func (h *Handler) issueToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
