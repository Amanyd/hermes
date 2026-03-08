package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "user_id"

func writeError(w http.ResponseWriter, status int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error":   message,
		"code":    code,
	})
}

func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || parts[0] != "Bearer" {
					writeError(w, http.StatusUnauthorized, "Invalid authorization format", "AUTH_INVALID")
					return
				}
				tokenString = parts[1]
			} else if t := r.URL.Query().Get("token"); t != "" {
				tokenString = t
			} else {
				writeError(w, http.StatusUnauthorized, "Missing authorization token", "AUTH_REQUIRED")
				return
			}

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
				return []byte(jwtSecret), nil
			}, jwt.WithValidMethods([]string{"HS256"}))
			if err != nil || !token.Valid {
				writeError(w, http.StatusUnauthorized, "Invalid or expired token", "AUTH_INVALID")
				return
			}

			subject, err := token.Claims.GetSubject()
			if err != nil || subject == "" {
				writeError(w, http.StatusUnauthorized, "Invalid token claims", "AUTH_INVALID")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) string {
	userID, _ := r.Context().Value(userIDKey).(string)
	return userID
}
