package http

import (
	"net/http"

	"pubsub-chat/internal/hub"
	"pubsub-chat/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func New(h *hub.Hub, lg *logger.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// WebSocket endpoint: /ws?room=<room>&name=<displayName>
	r.Get("/ws", h.ServeWS(lg))

	// Static test page
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	return r
}
