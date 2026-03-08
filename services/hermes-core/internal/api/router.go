package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(h *Handler, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3001"}, // Will change to frontend url
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", h.HealthCheck)

	r.Route("/api/v1", func(r chi.Router) {
		r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		// Public routes
		r.Post("/auth/register", h.Register)
		r.Options("/auth/register", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Post("/auth/login", h.Login)

		r.Get("/auth/callback/{provider}", h.OAuthCallback)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(JWTAuth(jwtSecret))

			r.Post("/relays", h.CreateRelay)
			r.Get("/relays", h.GetAllRelays)
			r.Get("/relays/{id}", h.GetRelay)
			r.Put("/relays/{id}", h.UpdateRelay)
			r.Put("/relays/{id}/actions", h.UpdateRelayActions)
			r.Delete("/relays/{id}", h.DeleteRelay)
			r.Get("/relays/{id}/executions", h.GetExecutions)
			r.Get("/executions/{executionId}/steps", h.GetExecutionSteps)
			r.Delete("/executions/{executionId}", h.DeleteExecution)
			r.Post("/relays/{id}/trigger", h.TriggerRelay)

			r.Post("/secrets", h.CreateSecret)
			r.Get("/secrets", h.ListSecrets)
			r.Delete("/secrets/{id}", h.DeleteSecret)

			r.Get("/connections/providers", h.AvailableProviders)
			r.Get("/connections", h.ListConnections)
			r.Get("/connections/{provider}/connect", h.ConnectProvider)
			r.Delete("/connections/{id}", h.DeleteConnection)
		})
	})
	return r
}
