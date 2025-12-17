package reverseproxy

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/komaldsukhani/reverseproxyexample/internal/memcache"
)

type ReverseProxy struct {
	TargetURL string
	Cache     *memcache.MemoryCache
}

func (p *ReverseProxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	//check if the request can be served from cache
	if r.Method == http.MethodGet {
		key := getCacheKey(r)
		cachedResp := p.Cache.Get(key)
		if cachedResp != nil {
			slog.Debug("Request served from the cache")
			rw.Write(cachedResp.Body)

			for h, vals := range cachedResp.Headers {
				for _, v := range vals {
					rw.Header().Add(h, v)
				}
			}

			return
		}
	}

	outreq, err := http.NewRequest(r.Method, p.TargetURL+r.URL.Path, r.Body)
	if err != nil {
		slog.Error("failed to create new http request", "error", err)

		http.Error(rw, "failed to handle request", http.StatusBadGateway)
		return
	}

	outreq.Header = r.Header.Clone()

	// remove hop-by-hop headers before sending to upstream
	removeHopByHopHeaders(outreq.Header)

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

	removeHopByHopHeaders(resp.Header)

	for h, vals := range resp.Header {
		for _, v := range vals {
			rw.Header().Add(h, v)
		}
	}

	rw.Write(body)

	//only cache get requests with 200 status code
	if r.Method == http.MethodGet && resp.StatusCode == http.StatusOK {
		record := memcache.Record{
			StatusCode: resp.StatusCode,
			Body:       bytes.Clone(body),
			Headers:    r.Header.Clone(),
		}

		key := getCacheKey(r)
		if err := p.Cache.Set(key, &record); err != nil {
			slog.Debug("failed to cache request", "error", err)
		}
	}

	slog.Debug("Successfully proxied the request")
}

func getCacheKey(r *http.Request) string {
	return r.Method + ":" + r.URL.String()
}

var hopByHopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

func removeHopByHopHeaders(header http.Header) {
	for _, vals := range header["Connection"] {
		for v := range strings.SplitSeq(vals, ",") {
			if v = strings.TrimSpace(v); v != "" {
				header.Del(v)
			}
		}

		for _, h := range hopByHopHeaders {
			header.Del(h)
		}
	}
}
