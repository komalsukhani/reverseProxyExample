package config

import "time"

const (
	DefaultUpstreamURL     = "http://example.com"
	DefaultListenPort      = 8080
	DefaultShutdownTimeout = 10 * time.Second

	DefaultCacheTTL           = 30 * time.Second
	DefaultMaxCacheSize       = 1 * 1024 * 1024
	DefaultMaxCacheRecordSize = 1 * 1024
)

type Config struct {
	LogLevel string
	Proxy    ProxyConfig
	Cache    CacheConfig
}

type ProxyConfig struct {
	ListenPort      int
	ShutdownTimeout time.Duration
	TargetURL       string
}

type CacheConfig struct {
	TTL           time.Duration
	MaxSize       int
	MaxRecordSize int
}

func (config *Config) SetDefaults() {
	switch config.LogLevel {
	case "debug", "info", "error", "warn":
	default:
		config.LogLevel = "info"
	}

	if config.Proxy.ListenPort == 0 {
		config.Proxy.ListenPort = DefaultListenPort
	}

	if config.Proxy.ShutdownTimeout == 0 {
		config.Proxy.ShutdownTimeout = DefaultShutdownTimeout
	}

	if config.Proxy.TargetURL == "" {
		config.Proxy.TargetURL = DefaultUpstreamURL
	}

	if config.Cache.TTL == 0 {
		config.Cache.TTL = DefaultCacheTTL
	}

	if config.Cache.MaxSize == 0 {
		config.Cache.MaxSize = DefaultMaxCacheSize
	}

	if config.Cache.MaxRecordSize == 0 {
		config.Cache.MaxRecordSize = DefaultMaxCacheRecordSize
	}
}
