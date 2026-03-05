package api

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/store"
	"github.com/go-chi/chi/v5"
)

// Redirects the user to the OAuth consent screen.
// GET /api/v1/connections/{provider}/connect
func (h *Handler) ConnectProvider(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	userID := GetUserID(r)

	provider, ok := h.oauthProviders[providerName]
	if !ok {
		h.respondError(w, http.StatusBadRequest, "Unsupported provider: "+providerName, "INVALID_PROVIDER")
		return
	}

	state, err := h.stateCodec.Encode(userID, providerName)
	if err != nil {
		h.logger.Error("failed to encode OAuth state", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to initiate connection", "INTERNAL_ERROR")
		return
	}

	authURL := provider.AuthURL(state)

	h.logger.Info("redirecting to OAuth provider",
		slog.String("provider", providerName),
		slog.String("user_id", userID),
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Handles the redirect from the OAuth provider.
// GET /api/v1/auth/callback/{provider}?code=...&state=...
func (h *Handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	code := r.URL.Query().Get("code")
	stateParam := r.URL.Query().Get("state")

	if code == "" {
		h.respondError(w, http.StatusBadRequest, "Missing authorization code", "MISSING_CODE")
		return
	}
	if stateParam == "" {
		h.respondError(w, http.StatusBadRequest, "Missing state parameter", "MISSING_STATE")
		return
	}

	userID, stateProvider, err := h.stateCodec.Decode(stateParam, 10*time.Minute)
	if err != nil {
		h.logger.Warn("invalid OAuth state",
			slog.String("provider", providerName),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusBadRequest, "Invalid or expired state parameter", "INVALID_STATE")
		return
	}

	if stateProvider != providerName {
		h.respondError(w, http.StatusBadRequest, "State provider mismatch", "INVALID_STATE")
		return
	}

	provider, ok := h.oauthProviders[providerName]
	if !ok {
		h.respondError(w, http.StatusBadRequest, "Unsupported provider", "INVALID_PROVIDER")
		return
	}

	tokens, err := provider.Exchange(r.Context(), code)
	if err != nil {
		h.logger.Error("OAuth token exchange failed",
			slog.String("provider", providerName),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusBadGateway, "Failed to exchange authorization code", "OAUTH_EXCHANGE_FAILED")
		return
	}

	// Store the connection (upsert: reconnecting updates the tokens).
	conn, err := h.connectionStore.Upsert(r.Context(), userID, providerName, tokens.Email,
		tokens.AccessToken, tokens.RefreshToken, "", tokens.Expiry)
	if err != nil {
		h.logger.Error("failed to store connection", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to save connection", "DB_ERROR")
		return
	}

	h.logger.Info("OAuth connection established",
		slog.String("provider", providerName),
		slog.String("user_id", userID),
		slog.String("account_email", tokens.Email),
	)

	h.respondSuccess(w, http.StatusOK, "Connection established", conn)
}

// Returns all OAuth connections for the authenticated user.
// GET /api/v1/connections
func (h *Handler) ListConnections(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	conns, err := h.connectionStore.ListByUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list connections", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to list connections", "DB_ERROR")
		return
	}
	h.respondSuccess(w, http.StatusOK, "", conns)
}

// Removes an OAuth connection.
// DELETE /api/v1/connections/{id}
func (h *Handler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	connID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	err := h.connectionStore.Delete(r.Context(), userID, connID)
	if err != nil {
		if errors.Is(err, store.ErrConnectionNotFound) {
			h.respondError(w, http.StatusNotFound, "Connection not found", "NOT_FOUND")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Failed to delete connection", "DB_ERROR")
		return
	}
	h.respondSuccess(w, http.StatusOK, "Connection removed", map[string]string{"deleted_id": connID})
}

// Returns which OAuth providers are configured on this instance.
// GET /api/v1/connections/providers
func (h *Handler) AvailableProviders(w http.ResponseWriter, r *http.Request) {
	providers := make([]string, 0, len(h.oauthProviders))
	for name := range h.oauthProviders {
		providers = append(providers, name)
	}
	h.respondSuccess(w, http.StatusOK, "", map[string]any{"providers": providers})
}
