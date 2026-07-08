package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	rlGetMaxRaw := os.Getenv("RATE_LIMIT_GET_MAX")
	rlPostMaxRaw := os.Getenv("RATE_LIMIT_POST_MAX")
	rlWindowSecRaw := os.Getenv("RATE_LIMIT_WINDOW_SEC")

	if redisAddr == "" || port == "" || allowedOrigins == "" ||
		rlGetMaxRaw == "" || rlPostMaxRaw == "" || rlWindowSecRaw == "" {
		return fmt.Errorf("main.run: some variables from env came empty (check REDIS_ADDR, SERVER_PORT, ALLOWED_ORIGINS, RATE_LIMIT_GET_MAX, RATE_LIMIT_POST_MAX, RATE_LIMIT_WINDOW_SEC)")
	}

	rlGetMax, err := strconv.Atoi(rlGetMaxRaw)
	if err != nil || rlGetMax <= 0 {
		return fmt.Errorf("main.run: RATE_LIMIT_GET_MAX must be a positive integer")
	}
	rlPostMax, err := strconv.Atoi(rlPostMaxRaw)
	if err != nil || rlPostMax <= 0 {
		return fmt.Errorf("main.run: RATE_LIMIT_POST_MAX must be a positive integer")
	}
	rlWindowSec, err := strconv.Atoi(rlWindowSecRaw)
	if err != nil || rlWindowSec <= 0 {
		return fmt.Errorf("main.run: RATE_LIMIT_WINDOW_SEC must be a positive integer")
	}
	rlWindow := time.Duration(rlWindowSec) * time.Second

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

	rlGet := middleware.RateLimit(redisClient, "napkin:rl:get:", rlGetMax, rlWindow)
	rlPost := middleware.RateLimit(redisClient, "napkin:rl:post:", rlPostMax, rlWindow)

	mux := http.NewServeMux()
	mux.Handle("POST /api/save", rlPost(http.HandlerFunc(h.Save)))
	mux.Handle("GET /{code}", rlGet(http.HandlerFunc(h.Get)))
	mux.Handle("GET /{code}/ws", rlGet(http.HandlerFunc(h.WebSocket)))

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
