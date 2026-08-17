package server

import (
	"context"
	"errors"
	"fmt"
	"mmktestbasisByDGanichev/internal/auth"
	"mmktestbasisByDGanichev/internal/comment"
	"mmktestbasisByDGanichev/internal/config"
	"mmktestbasisByDGanichev/internal/history"
	"mmktestbasisByDGanichev/internal/middleware"
	"mmktestbasisByDGanichev/internal/stats"
	"mmktestbasisByDGanichev/internal/task"
	"mmktestbasisByDGanichev/internal/team"
	"mmktestbasisByDGanichev/internal/user"
	"net/http"
	"time"
)

type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

func New(
	cfg config.HTTPConfig,
	userHandler *user.Handler,
	authHandler *auth.Handler,
	authMiddleware *middleware.Auth,
	teamHandler *team.Handler,
	taskHandler *task.Handler,
	commentHandler *comment.Handler,
	historyHandler *history.Handler,
	statsHandler *stats.Handler,
) *Server {
	mux := http.NewServeMux()
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)

	s := &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           rateLimiter.Limit(mux),
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
			MaxHeaderBytes:    1 << 20,
		},
		shutdownTimeout: cfg.ShutdownTimeout,
	}

	s.registerRoutes(
		mux,
		userHandler,
		authHandler,
		authMiddleware,
		teamHandler,
		taskHandler,
		commentHandler,
		historyHandler,
		statsHandler,
	)

	return s
}

func (s *Server) registerRoutes(
	mux *http.ServeMux,
	userHandler *user.Handler,
	authHandler *auth.Handler,
	authMiddleware *middleware.Auth,
	teamHandler *team.Handler,
	taskHandler *task.Handler,
	commentHandler *comment.Handler,
	historyHandler *history.Handler,
	statsHandler *stats.Handler,
) {
	mux.HandleFunc("GET /health", s.handleHealth)

	mux.HandleFunc("POST /api/v1/register", userHandler.Register)

	mux.Handle(
		"POST /api/v1/teams",
		authMiddleware.Authenticate(http.HandlerFunc(teamHandler.Create)),
	)

	mux.Handle("GET /api/v1/teams", authMiddleware.Authenticate(http.HandlerFunc(teamHandler.List)))

	mux.Handle(
		"POST /api/v1/teams/{id}/invite",
		authMiddleware.Authenticate(http.HandlerFunc(teamHandler.Invite)),
	)

	mux.Handle(
		"PATCH /api/v1/teams/{id}/members/{user_id}",
		authMiddleware.Authenticate(http.HandlerFunc(teamHandler.UpdateMemberRole)),
	)

	mux.HandleFunc("POST /api/v1/login", authHandler.Login)

	mux.Handle(
		"POST /api/v1/tasks",
		authMiddleware.Authenticate(http.HandlerFunc(taskHandler.Create)),
	)

	mux.Handle("GET /api/v1/tasks", authMiddleware.Authenticate(http.HandlerFunc(taskHandler.List)))

	mux.Handle(
		"PUT /api/v1/tasks/{id}",
		authMiddleware.Authenticate(http.HandlerFunc(taskHandler.Update)),
	)

	mux.Handle(
		"POST /api/v1/tasks/{id}/comments",
		authMiddleware.Authenticate(http.HandlerFunc(commentHandler.Create)),
	)

	mux.Handle(
		"GET /api/v1/tasks/{id}/comments",
		authMiddleware.Authenticate(http.HandlerFunc(commentHandler.List)),
	)

	mux.Handle(
		"GET /api/v1/tasks/{id}/history",
		authMiddleware.Authenticate(http.HandlerFunc(historyHandler.List)),
	)

	mux.Handle(
		"GET /api/v1/teams/{team_id}/stats",
		authMiddleware.Authenticate(http.HandlerFunc(statsHandler.Report)),
	)
}

func (s *Server) handleHealth(
	w http.ResponseWriter,
	_ *http.Request,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.listenAndServe()
	}()

	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		if err := <-errCh; err != nil {
			return err
		}

		return nil
	}
}

func (s *Server) listenAndServe() error {
	err := s.httpServer.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}
