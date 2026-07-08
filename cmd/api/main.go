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

	"github.com/fernandesenzo/napkin/internal/infra"
	"github.com/fernandesenzo/napkin/internal/logger"
	"github.com/fernandesenzo/napkin/internal/manager"
	"github.com/fernandesenzo/napkin/internal/middleware"
	"github.com/fernandesenzo/napkin/internal/napkin/handler"
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
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPswd := os.Getenv("REDIS_PASSWORD")
	port := os.Getenv("SERVER_PORT")
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

	if redisAddr == "" || port == "" || allowedOrigins == "" {
		return fmt.Errorf("main.run: some variables from env came empty")
	}

	redisClient, err := infra.NewRedisClient(redisAddr, redisPswd)
	if err != nil {
		return fmt.Errorf("main.run: redis connection failed: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("main.run: failed to close redis", "err", err)
		} else {
			slog.Info("redis connection closed gracefully")
		}
	}()
	slog.Info("connected to redis succesfully")

	repo := repository.NewRedisRepository(redisClient)
	svc := service.New(repo)
	manager := manager.New(svc)
	h := handler.New(svc, manager)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/save", h.Save)
	mux.HandleFunc("GET /{code}", h.Get)
	mux.HandleFunc("GET /{code}/ws", h.WebSocket)

	var handlerStack http.Handler = mux
	handlerStack = middleware.AccessLog(handlerStack)
	handlerStack = middleware.ApplyHeaders(allowedOrigins)(handlerStack)
	handlerStack = middleware.InjectReqID(handlerStack)
	handlerStack = middleware.Recover(handlerStack)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handlerStack,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("server starting", "port", port)
		serverErrors <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		slog.Info("shutting down OS signal received")

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
	}
	return nil
}
