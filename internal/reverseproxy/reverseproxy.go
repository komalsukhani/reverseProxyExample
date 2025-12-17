package reverseproxy

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/komaldsukhani/reverseproxyexample/internal/memcache"
)

type ReverseProxy struct {
	TargetURL string
	Cache     *memcache.MemoryCache
}

func (p *ReverseProxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// Check if the request can be served from cache.
	if r.Method == http.MethodGet {
		key := getCacheKey(r)
		cachedResp := p.Cache.Get(key)
		if cachedResp != nil {
			slog.Debug("Request served from the cache", "key", key, "status", cachedResp.StatusCode)

			for h, vals := range cachedResp.Headers {
				for _, v := range vals {
					rw.Header().Add(h, v)
				}
			}

			rw.WriteHeader(cachedResp.StatusCode)

			_, err := rw.Write(cachedResp.Body)
			if err != nil {
				slog.Error("failed to write cached response body", "error", err)
			} else {
				return
			}
		}
		// cache miss
		if cachedResp == nil {
			slog.Debug("Cache miss", "key", key)
		}
	}

	outreq, err := http.NewRequestWithContext(r.Context(), r.Method, "", r.Body)
	if err != nil {
		slog.Error("failed to create new http request", "error", err)

		http.Error(rw, "failed to handle request", http.StatusBadGateway)
		return
	}

	if outreq.URL, err = joinURL(r.URL, p.TargetURL); err != nil {
		slog.Error("failed to join target url and request path", "error", err)

		http.Error(rw, "failed to handle request", http.StatusBadGateway)
		return
	}

	outreq.Header = r.Header.Clone()
	slog.Debug("Prepared outbound request", "method", outreq.Method, "url", outreq.URL.String())

	// Remove hop-by-hop headers before sending to upstream
	removeHopByHopHeaders(outreq.Header)

	resp, err := http.DefaultClient.Do(outreq)
	if err != nil {
		slog.Error("request to upstream failed", "error", err)

		http.Error(rw, "failed to handle request", http.StatusBadGateway)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
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

	rw.WriteHeader(resp.StatusCode)

	if _, err := rw.Write(body); err != nil {
		slog.Error("failed to write response body", "error", err)

		return
	}

	// Caching only GET requests with 200 OK response
	if r.Method == http.MethodGet && resp.StatusCode == http.StatusOK {
		record := memcache.Record{
			StatusCode: resp.StatusCode,
			Body:       bytes.Clone(body),
			Headers:    resp.Header.Clone(),
		}

		key := getCacheKey(r)
		if err := p.Cache.Set(key, &record); err != nil {
			slog.Debug("failed to cache request", "error", err)
		} else {
			slog.Debug("Request cached", "key", key, "size", record.Calsize())
		}
	}

	slog.Debug("Successfully proxied the request")
}

func getCacheKey(r *http.Request) string {
	return r.Method + ":" + r.URL.String()
}

func joinURL(req *url.URL, targetURL string) (*url.URL, error) {
	var joinedURL url.URL

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	joinedURL.Host = parsedURL.Host
	joinedURL.Scheme = parsedURL.Scheme

	if parsedURL.RawPath == "" && req.RawPath == "" {
		joinedURL.Path = strings.TrimRight(parsedURL.Path, "/") + "/" + strings.TrimLeft(req.Path, "/")
		joinedURL.RawPath = ""
	} else {
		joinedURL.Path = strings.TrimRight(parsedURL.Path, "/") + "/" + strings.TrimLeft(req.Path, "/")

		escapedTargetPath := parsedURL.EscapedPath()
		escapedReqPath := req.EscapedPath()

		joinedURL.RawPath = strings.TrimRight(escapedTargetPath, "/") + "/" + strings.TrimLeft(escapedReqPath, "/")
	}

	if parsedURL.RawQuery == "" || req.RawQuery == "" {
		joinedURL.RawQuery = parsedURL.RawQuery + req.RawQuery
	} else {
		joinedURL.RawQuery = parsedURL.RawQuery + "&" + req.RawQuery
	}

	return &joinedURL, nil
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

// RFC-conformant removal of hop-by-hop headers. We look for header
// values in the "Connection" header that name additional single headers
// to drop, trim & drop them, then drop known hop-by-hop header names.
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
