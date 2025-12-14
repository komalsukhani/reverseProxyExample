package internal

import (
	"io"
	"log/slog"
	"net/http"
)

type ReverseProxy struct {
	TargetURL string
}

func (p *ReverseProxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	outreq, err := http.NewRequest(r.Method, p.TargetURL+r.URL.Path, r.Body)
	if err != nil {
		slog.Error("failed to create new http request", "error", err)

		http.Error(rw, "failed to handle request", http.StatusBadGateway)
		return
	}

	resp, err := http.DefaultClient.Do(outreq)
	if err != nil {
		slog.Error("request to upstream failed", "error", err)

		http.Error(rw, "failed to handle request", http.StatusBadGateway)
		return
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		slog.Error("failed to read response body of upstream request", "error", err)

		http.Error(rw, "failed to handle request", http.StatusBadGateway)
		return
	}

	for h, vals := range resp.Header {
		for _, v := range vals {
			rw.Header().Add(h, v)
		}
	}

	rw.Write(body)

	slog.Debug("Successfully proxied the request")
}
