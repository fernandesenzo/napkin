package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fernandesenzo/napkin/internal/config"
	healthHandler "github.com/fernandesenzo/napkin/internal/health/handler"
	"github.com/fernandesenzo/napkin/internal/infra"
	"github.com/fernandesenzo/napkin/internal/logger"
	"github.com/fernandesenzo/napkin/internal/manager"
	"github.com/fernandesenzo/napkin/internal/middleware"
	napkinHandler "github.com/fernandesenzo/napkin/internal/napkin/handler"
	"github.com/fernandesenzo/napkin/internal/napkin/repository"
	"github.com/fernandesenzo/napkin/internal/napkin/service"
	"github.com/joho/godotenv"
)

func main() {
	logger.Setup()
	if err := run(); err != nil {
		slog.Error("application failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		slog.Info("could not read .env file. assuming they are already injected")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	redisClient, err := infra.NewRedisClient(cfg.RedisAddress, cfg.RedisPassword)
	if err != nil {
		return fmt.Errorf("main.run: redis connection failed: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("main.run: failed to close redis", "err", err)
		} else {
			slog.Info("main.run: redis connection closed gracefully")
		}
	}()
	slog.Info("main.run: connected to redis succesfully")

	repo := repository.NewRedisRepository(redisClient)
	svc := service.New(repo, service.Config{
		DefaultTTL:       cfg.DefaultTTL,
		CodeLength:       cfg.CodeLength,
		MaxContentLength: cfg.MaxContentLength,
	})
	manager := manager.New(svc)
	h := napkinHandler.New(svc, manager, napkinHandler.Config{
		MaxContentLength: cfg.MaxContentLength,
		CodeLength:       cfg.CodeLength,
		AllowedOrigins:   cfg.AllowedOrigins,
	})
	healthHandler := healthHandler.New(redisClient)

	rlGet := middleware.RateLimit(redisClient, "napkin:rl:get:", cfg.RateLimitGetMax, cfg.RateLimitWindow)
	rlPost := middleware.RateLimit(redisClient, "napkin:rl:post:", cfg.RateLimitPostMax, cfg.RateLimitWindow)

	mux := http.NewServeMux()
	mux.Handle("POST /save", middleware.BodyLimit(4096)(rlPost(http.HandlerFunc(h.Save))))
	mux.Handle("GET /{code}", rlGet(http.HandlerFunc(h.Get)))
	mux.Handle("GET /{code}/ws", rlGet(http.HandlerFunc(h.WebSocket)))
	mux.HandleFunc("GET /health", healthHandler.Get)

	var handlerStack http.Handler = mux
	handlerStack = middleware.AccessLog(handlerStack)
	handlerStack = middleware.ApplyHeaders(cfg.AllowedOrigins)(handlerStack)
	handlerStack = middleware.InjectReqID(handlerStack)
	handlerStack = middleware.Recover(handlerStack)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      handlerStack,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("main.run: server starting", "port", cfg.ServerPort)
		serverErrors <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		slog.Info("main.run: shutting down OS signal received")

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
	}
	return nil
}
