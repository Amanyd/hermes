package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "user_id"

func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"success":false,"error":"Missing authorization header","code":"AUTH_REQUIRED"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"success":false,"error":"Invalid authorization format","code":"AUTH_INVALID"}`, http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
				return []byte(jwtSecret), nil
			}, jwt.WithValidMethods([]string{"HS256"}))
			if err != nil || !token.Valid {
				http.Error(w, `{"success":false,"error":"Invalid or expired token","code":"AUTH_INVALID"}`, http.StatusUnauthorized)
				return
			}

			subject, err := token.Claims.GetSubject()
			if err != nil || subject == "" {
				http.Error(w, `{"success":false,"error":"Invalid token claims","code":"AUTH_INVALID"}`, http.StatusUnauthorized)
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