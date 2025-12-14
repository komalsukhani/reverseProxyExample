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
	TargetURL = "http://example.com"
	TTL       = 30 * time.Second
	Port      = 8080
)

func main() {
	addr := fmt.Sprintf(":%d", Port)

	p := internal.ReverseProxy{
		TargetURL: TargetURL,
		Cache:     internal.NewMemoryCache(TTL),
	}

	slog.Info("Started server", "port", Port)
	log.Fatal(http.ListenAndServe(addr, &p))
}
