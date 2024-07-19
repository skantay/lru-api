package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/skantay/lru-api/internal/api"
	"github.com/skantay/lru-api/internal/cache"
	"github.com/skantay/lru-api/pkg/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error(err.Error())

		os.Exit(1)
	}

	var level slog.Level

	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "info":
		level = slog.LevelInfo
	case "error":
		level = slog.LevelError
	}

	log := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: level,
			},
		),
	)

	cache, err := cache.New(
		cfg.CacheSize,
		time.Duration(cfg.DefaultCacheTTL)*time.Second,
		log,
	)
	if err != nil {
		log.Error(err.Error())

		os.Exit(1)
	}

	handler := api.New(cache, log)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.HTTPPort),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("ListenAndServe: " + err.Error())

			select {
			case done <- syscall.SIGTERM:
			default:
			}
		}
	}()

	log.Info("Starting server", "port", cfg.HTTPPort)
	now := time.Now()
	log.Info("Signal is captured", "signal", <-done)
	log.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server Shutdown Failed", "error", err.Error())
	} else {
		log.Info("Server exited properly", "shutdown duration", time.Since(now))
	}
}
