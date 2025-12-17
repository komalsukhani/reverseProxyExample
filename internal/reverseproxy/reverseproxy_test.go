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

	"github.com/komaldsukhani/reverseproxyexample/internal/memcache"
	"github.com/matryer/is"
)

func TestCachedRequest(t *testing.T) {
	eval := is.New(t)

	var upstreamCalls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstreamCalls, 1)
		w.Header().Add("Upstream", "true")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	}))
	defer srv.Close()

	rproxy := ReverseProxy{
		TargetURL: srv.URL,
		Cache:     memcache.NewMemoryCache(5*time.Minute, 1*1024*1024, 1024),
	}

	proxysrv := httptest.NewServer(&rproxy)
	defer proxysrv.Close()

	resp1, err := http.Get(proxysrv.URL)
	eval.NoErr(err)

	var resp2 *http.Response
	resp2, err = http.Get(proxysrv.URL)
	eval.NoErr(err)

	var body1, body2 []byte
	body1, err = io.ReadAll(resp1.Body)
	defer resp1.Body.Close()
	eval.NoErr(err)

	body2, err = io.ReadAll(resp2.Body)
	defer resp2.Body.Close()
	eval.NoErr(err)
	eval.True(compareHeaders(t, resp1.Header, resp2.Header))

	eval.Equal(resp1.StatusCode, resp2.StatusCode)
	eval.Equal(body1, body2)

	eval.Equal(atomic.LoadInt32(&upstreamCalls), int32(1))
}

func compareHeaders(t *testing.T, h1, h2 http.Header) bool {
	if !(len(h1) == len(h2)) {
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
		w.Write([]byte("Hello World"))
	}))
	defer srv.Close()

	rproxy := ReverseProxy{
		TargetURL: srv.URL,
		Cache:     memcache.NewMemoryCache(5*time.Minute, 1*1024*1024, 1024),
	}

	proxysrv := httptest.NewServer(&rproxy)
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
		w.Write([]byte("Hello World"))
	}))
	defer srv.Close()

	rproxy := ReverseProxy{
		TargetURL: srv.URL,
		Cache:     memcache.NewMemoryCache(5*time.Minute, 1*1024*1024, 1024),
	}

	proxysrv := httptest.NewServer(&rproxy)
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
		w.Write([]byte("Hello World"))
	}))
	defer srv.Close()

	rproxy := ReverseProxy{
		TargetURL: srv.URL,
		Cache:     memcache.NewMemoryCache(30*time.Second, 1*1024*1024, 1024),
	}

	proxysrv := httptest.NewServer(&rproxy)
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

	for _, tc := range testcases {
		reqURL, err := url.Parse(tc.reqPath)
		eval.NoErr(err)
		u := joinURL(reqURL, tc.targetURL)

		eval.Equal(u.String(), tc.wantFinal)
	}
}
