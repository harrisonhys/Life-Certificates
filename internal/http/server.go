package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/swaggo/http-swagger"

	"life-certificates/internal/config"
	handlers "life-certificates/internal/http/handler"
	custommiddleware "life-certificates/internal/http/middleware"
	"life-certificates/internal/http/response"
)

// Server wraps the HTTP server lifecycle.
type Server struct {
	httpServer *http.Server
}

// NewServer assembles the HTTP router and dependencies.
func NewServer(cfg *config.Config, participantHandler *handlers.ParticipantHandler, memberHandler *handlers.MemberHandler, lifeHandler *handlers.LifeCertificateHandler) *Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		response.Success(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Group(func(r chi.Router) {
		r.Use(custommiddleware.BasicAuth(cfg.Auth.Username, cfg.Auth.Password))

		r.Route("/participants", func(r chi.Router) {
			r.Get("/", participantHandler.List)
			r.Get("/{participant_id}", participantHandler.Get)
			r.Put("/{participant_id}", participantHandler.Update)
			r.Delete("/{participant_id}", participantHandler.Delete)
			r.Post("/register", participantHandler.Register)
		})

		r.Route("/members", func(r chi.Router) {
			r.Post("/", memberHandler.Create)
			r.Get("/", memberHandler.List)
			r.Get("/{member_id}", memberHandler.Get)
			r.Put("/{member_id}", memberHandler.Update)
			r.Delete("/{member_id}", memberHandler.Delete)
		})

		r.Route("/life-certificate", func(r chi.Router) {
			r.Post("/verify", lifeHandler.Verify)
			r.Get("/status/{participant_id}", lifeHandler.LatestStatus)
		})

		r.Get("/swagger/*", httpSwagger.Handler())
	})

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	return &Server{httpServer: httpServer}
}

// Start begins serving HTTP traffic.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown performs a graceful server shutdown.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
