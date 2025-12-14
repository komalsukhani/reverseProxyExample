package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/komaldsukhani/reverseproxyexample/internal"
)

const (
	UpstreamURL = "http://example.com"
	Port        = 8080

	CacheTTL     = 30 * time.Second
	MaxCacheSize = 1 * 1024 * 1024
)

func main() {
	addr := fmt.Sprintf(":%d", Port)

	p := internal.ReverseProxy{
		TargetURL: UpstreamURL,
		Cache:     internal.NewMemoryCache(CacheTTL, MaxCacheSize),
	}

	slog.Info("Started server", "port", Port)
	log.Fatal(http.ListenAndServe(addr, &p))
}
