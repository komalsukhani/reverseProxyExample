package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/komaldsukhani/reverseproxyexample/internal/config"
	rproxy "github.com/komaldsukhani/reverseproxyexample/internal/reverseproxy"
)

func main() {
	var config config.Config

	if err := readConfig(&config); err != nil {
		slog.Error("failed to read config", "err", err)

		return
	}

	setupLogger(&config)
	slog.Debug("Configured logger", "loglevel", config.LogLevel)

	addr := fmt.Sprintf(":%d", config.Proxy.Server.ListenPort)

	p := rproxy.New(&config)

	srv := http.Server{
		Addr:         addr,
		Handler:      p,
		ReadTimeout:  config.Proxy.Server.ReadTimeout,
		WriteTimeout: config.Proxy.Server.WriteTimeout,
		IdleTimeout:  config.Proxy.Server.IdleTimeout,
	}

	go func() {
		slog.Info("Started server", "addr", srv.Addr)

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	gracefulShutdown(&srv, &config)
}

func gracefulShutdown(srv *http.Server, config *config.Config) {
	var quit = make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	slog.Info("Shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), config.Proxy.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown the server: %v", err)
	}
}

func readConfig(config *config.Config) error {
	if err := envconfig.Process("", config); err != nil {
		slog.Error("failed to read config", "error", err)

		return err
	}

	config.SetDefaults()

	return nil
}

func setupLogger(config *config.Config) {
	var level slog.Level

	switch config.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "info":
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
}
