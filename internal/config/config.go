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

	DefaultTransportMaxIdleConnections  = 100
	DefaultTransportMaxIdleConnsPerHost = 20
	DefaultTransportIdleConnTimeout     = 90 * time.Second
	DefaultTransportDialTimeout         = 5 * time.Second
)

type Config struct {
	LogLevel string
	Proxy    ProxyConfig
	Cache    CacheConfig
}

type ProxyConfig struct {
	Server    HTTPServerConfig
	TargetURL string
	Transport TransportConfig
}

type TransportConfig struct {
	MaxIdleConnections  int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	DialTimeout         time.Duration
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

	if config.Proxy.Server.ListenPort == 0 {
		config.Proxy.Server.ListenPort = DefaultListenPort
	}

	if config.Proxy.Server.ShutdownTimeout == 0 {
		config.Proxy.Server.ShutdownTimeout = DefaultShutdownTimeout
	}

	if config.Proxy.Server.ReadTimeout == 0 {
		config.Proxy.Server.ReadTimeout = DefaultReadTimeout
	}

	if config.Proxy.Server.WriteTimeout == 0 {
		config.Proxy.Server.WriteTimeout = DefaultWriteTimeout
	}

	if config.Proxy.Server.IdleTimeout == 0 {
		config.Proxy.Server.IdleTimeout = DefaultIdleTimeout
	}

	if config.Proxy.Transport.MaxIdleConnections == 0 {
		config.Proxy.Transport.MaxIdleConnections = DefaultTransportMaxIdleConnections
	}

	if config.Proxy.Transport.MaxIdleConnsPerHost == 0 {
		config.Proxy.Transport.MaxIdleConnsPerHost = DefaultTransportMaxIdleConnsPerHost
	}

	if config.Proxy.Transport.IdleConnTimeout == 0 {
		config.Proxy.Transport.IdleConnTimeout = DefaultTransportIdleConnTimeout
	}

	if config.Proxy.Transport.DialTimeout == 0 {
		config.Proxy.Transport.DialTimeout = DefaultTransportDialTimeout
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
