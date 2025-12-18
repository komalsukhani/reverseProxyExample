package reverseproxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/komaldsukhani/reverseproxyexample/internal/config"
	"github.com/matryer/is"
)

func TestCachedRequest(t *testing.T) {
	eval := is.New(t)

	var upstreamCalls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstreamCalls, 1)
		w.Header().Add("Upstream", "true")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	rproxy := New(&config.Config{
		Proxy: config.ProxyConfig{
			TargetURL: srv.URL,
		},
		Cache: config.CacheConfig{
			TTL:           5 * time.Minute,
			MaxSize:       1 * 1024 * 1024,
			MaxRecordSize: 1024,
		},
	})

	proxysrv := httptest.NewServer(rproxy)
	defer proxysrv.Close()

	resp1, err := http.Get(proxysrv.URL)
	eval.NoErr(err)

	var resp2 *http.Response
	resp2, err = http.Get(proxysrv.URL)
	eval.NoErr(err)

	var body1, body2 []byte
	body1, err = io.ReadAll(resp1.Body)
	defer func() { _ = resp1.Body.Close() }()
	eval.NoErr(err)

	body2, err = io.ReadAll(resp2.Body)
	defer func() { _ = resp2.Body.Close() }()
	eval.NoErr(err)
	eval.True(compareHeaders(t, resp1.Header, resp2.Header))

	eval.Equal(resp1.StatusCode, resp2.StatusCode)
	eval.Equal(body1, body2)

	eval.Equal(atomic.LoadInt32(&upstreamCalls), int32(1))
}

func compareHeaders(t *testing.T, h1, h2 http.Header) bool {
	if len(h1) != len(h2) {
		return false
	}

	for h, vals := range h1 {
		for _, v := range vals {
			if !(slices.Contains(h2[h], v)) {
				t.Logf("Header mismatch for %s: %v vs %v", h, h1[h], h2[h])
				return false
			}
		}
	}

	return true
}

func TestCacheNonSupportedMethods(t *testing.T) {
	eval := is.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	rproxy := New(&config.Config{
		Proxy: config.ProxyConfig{
			TargetURL: srv.URL,
		},
	})

	proxysrv := httptest.NewServer(rproxy)
	defer proxysrv.Close()

	_, err := http.Post(proxysrv.URL, "", nil)
	eval.NoErr(err)

	eval.True(rproxy.Cache.Count() == 0)
}

func TestCachedNonSupportedResponseCode(t *testing.T) {
	eval := is.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/protected" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	rproxy := New(&config.Config{
		Proxy: config.ProxyConfig{
			TargetURL: srv.URL,
		},
	})

	proxysrv := httptest.NewServer(rproxy)
	defer proxysrv.Close()

	_, err := http.Get(proxysrv.URL + "/protected")
	eval.NoErr(err)

	eval.True(rproxy.Cache.Count() == 0)
}

func TestCacheTTL(t *testing.T) {
	eval := is.New(t)

	var upstreamCalls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstreamCalls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	rproxy := New(&config.Config{
		Proxy: config.ProxyConfig{
			TargetURL: srv.URL,
		},
		Cache: config.CacheConfig{
			TTL: 30 * time.Second,
		},
	})

	proxysrv := httptest.NewServer(rproxy)
	defer proxysrv.Close()

	_, err := http.Get(proxysrv.URL)
	eval.NoErr(err)

	time.Sleep(30 * time.Second)

	_, err = http.Get(proxysrv.URL)
	eval.NoErr(err)

	eval.Equal(atomic.LoadInt32(&upstreamCalls), int32(2))
}

func TestJoinURL(t *testing.T) {
	eval := is.New(t)

	testcases := map[string]struct {
		targetURL string
		reqPath   string
		wantFinal string
	}{
		"trailing slash in target url":                                   {"http://example.com/api/", "users", "http://example.com/api/users"},
		"trailing slash in target url and leading slash in request path": {"http://example.com/api/", "/users", "http://example.com/api/users"},
		"without query parameters":                                       {"http://example.com/api", "/users", "http://example.com/api/users"},
		"query parameters in target url":                                 {"http://example.com/api?a=test", "users", "http://example.com/api/users?a=test"},
		"query parameters in request path and target url":                {"http://example.com/api?a=test", "users?b=test", "http://example.com/api/users?a=test&b=test"},
		"query parameters in request path":                               {"http://example.com/api/", "/users?a=test", "http://example.com/api/users?a=test"},
		"request path with special characters":                           {"http://example.com/api", "foo%2Fbar", "http://example.com/api/foo%2Fbar"},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			reqURL, err := url.Parse(tc.reqPath)
			eval.NoErr(err)

			u, err := joinURL(reqURL, tc.targetURL)
			eval.NoErr(err)

			eval.Equal(u.String(), tc.wantFinal)
		})
	}
}

func TestCanCacheRequest(t *testing.T) {
	eval := is.New(t)

	testcases := map[string]struct {
		method       string
		headers      http.Header
		statusCode   int
		wantCanCache bool
	}{
		"GET request with 200 OK":                 {method: http.MethodGet, statusCode: http.StatusOK, wantCanCache: true},
		"HEAD request with 200 OK":                {method: http.MethodHead, statusCode: http.StatusOK, wantCanCache: true},
		"GET request with 404 Not Found":          {method: http.MethodGet, statusCode: http.StatusNotFound, wantCanCache: false},
		"POST request with 200 OK":                {method: http.MethodPost, statusCode: http.StatusOK, wantCanCache: false},
		"GET request with no-store cache control": {method: http.MethodGet, headers: http.Header{"Cache-Control": []string{"no-store"}}, statusCode: http.StatusOK, wantCanCache: false},
		"GET request with no-cache cache control": {method: http.MethodGet, headers: http.Header{"Cache-Control": []string{"no-cache"}}, statusCode: http.StatusOK, wantCanCache: false},
		"Get request with private cache control":  {method: http.MethodGet, headers: http.Header{"Cache-Control": []string{"private"}}, statusCode: http.StatusOK, wantCanCache: false},
		"GET request with Authorization header":   {method: http.MethodGet, headers: http.Header{"Authorization": []string{"Bearer token"}}, statusCode: http.StatusOK, wantCanCache: true},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, "http://example.com", nil)
			eval.NoErr(err)

			resp := &http.Response{
				StatusCode: tc.statusCode,
				Header:     make(http.Header),
			}

			if tc.headers != nil {
				resp.Header = tc.headers.Clone()
			}

			canCache := canCacheRequest(req, resp)
			eval.Equal(canCache, tc.wantCanCache)
		})
	}
}
