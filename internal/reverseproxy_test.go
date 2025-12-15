package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestCachedRequest(t *testing.T) {
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
		Cache:     NewMemoryCache(5*time.Minute, 1*1024*1024, 1024),
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

	eval.Equal(resp1.StatusCode, resp2.StatusCode)
	eval.Equal(body1, body2)

	eval.Equal(atomic.LoadInt32(&upstreamCalls), int32(1))
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
		Cache:     NewMemoryCache(5*time.Minute, 1*1024*1024, 1024),
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
		Cache:     NewMemoryCache(5*time.Minute, 1*1024*1024, 1024),
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
		Cache:     NewMemoryCache(30*time.Second, 1*1024*1024, 1024),
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
