package config

import "time"

const (
	DefaultUpstreamURL = "http://httpbin.org"

	DefaultListenPort      = 8080
	DefaultShutdownTimeout = 10 * time.Second
	DefaultReadTimeout     = 10 * time.Second
	DefaultWriteTimeout    = 10 * time.Second
	DefaultIdleTimeout     = 120 * time.Second

	DefaultCacheTTL           = 1 * time.Minute
	DefaultMaxCacheSize       = 1 * 1024 * 1024
	DefaultMaxCacheRecordSize = 1 * 1024
)

type Config struct {
	LogLevel string
	Proxy    ProxyConfig
	Cache    CacheConfig
}

type ProxyConfig struct {
	ServerConfig HTTPServerConfig
	TargetURL    string
}

type HTTPServerConfig struct {
	ListenPort int

	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
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

	if config.Proxy.ServerConfig.ListenPort == 0 {
		config.Proxy.ServerConfig.ListenPort = DefaultListenPort
	}

	if config.Proxy.ServerConfig.ShutdownTimeout == 0 {
		config.Proxy.ServerConfig.ShutdownTimeout = DefaultShutdownTimeout
	}

	if config.Proxy.ServerConfig.ReadTimeout == 0 {
		config.Proxy.ServerConfig.ReadTimeout = DefaultReadTimeout
	}

	if config.Proxy.ServerConfig.WriteTimeout == 0 {
		config.Proxy.ServerConfig.WriteTimeout = DefaultWriteTimeout
	}

	if config.Proxy.ServerConfig.IdleTimeout == 0 {
		config.Proxy.ServerConfig.IdleTimeout = DefaultIdleTimeout
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
