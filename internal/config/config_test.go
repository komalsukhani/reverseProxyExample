package config

import (
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/matryer/is"
)

func TestProxyTargetURLFromEnv(t *testing.T) {
	eval := is.New(t)

	// set an explicit upstream URL via env var
	const want = "http://example-upstream.test"
	t.Setenv("PROXY_TARGETURL", want)

	var cfg Config
	err := envconfig.Process("", &cfg)
	eval.NoErr(err)

	cfg.SetDefaults()

	eval.Equal(cfg.Proxy.TargetURL, want)
}
