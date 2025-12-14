package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/komaldsukhani/reverseproxyexample/internal"
)

func main() {
	// TODO: make port configurable
	port := 8080
	addr := fmt.Sprintf(":%d", port)

	p := internal.ReverseProxy{
		TargetURL: "http://example.com",
	}

	slog.Info("Started server", "port", port)
	log.Fatal(http.ListenAndServe(addr, &p))
}
