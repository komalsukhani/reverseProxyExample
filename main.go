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
	"time"

	"github.com/komaldsukhani/reverseproxyexample/internal"
)

const (
	UpstreamURL     = "http://example.com"
	Port            = 8080
	ShutdownTimeout = 10 * time.Second

	CacheTTL           = 30 * time.Second
	MaxCacheSize       = 1 * 1024 * 1024
	MaxCacheRecordSize = 1 * 1024
)

func main() {
	addr := fmt.Sprintf(":%d", Port)

	p := internal.ReverseProxy{
		TargetURL: UpstreamURL,
		Cache:     internal.NewMemoryCache(CacheTTL, MaxCacheSize, MaxCacheRecordSize),
	}

	srv := http.Server{
		Addr: addr,
	}

	go func() {
		slog.Info("Started server", "port", Port)

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}

		log.Fatal(http.ListenAndServe(addr, &p))
	}()

	gracefulShutdown(&srv)
}

func gracefulShutdown(srv *http.Server) {
	var quit = make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	slog.Info("Shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown the server: %v", err)
	}
}
