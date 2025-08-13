package httpx

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter assembles the API routes.
func NewRouter(handler Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(middleware.StripSlashes)

	// Health check endpoint
	r.Get("/healthz", handler.Health)

	// Management endpoints
	r.Route("/v1", func(r chi.Router) {
		r.Route("/models", func(r chi.Router) {
			r.Get("/", handler.ListModels)
			r.Post("/", handler.CreateModel)
		})

		// Unified proxy endpoint compatible with OpenAI-style paths
		r.Post("/chat/completions", handler.ProxyChatCompletions)
	})

	return r
}


