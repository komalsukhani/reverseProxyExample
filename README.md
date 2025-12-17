# reverseProxyExample

## How to run

- Run the server locally:

```bash
go run ./main.go
```

- Run the server with example environment variables:

```bash
LOGLEVEL=debug PROXY_LISTENPORT=9090 CACHE_TTL=30s go run ./main.go
```

## How to run tests

- Run the unit test suite:

```bash
go test ./...
```

## Configuration (environment variables)

The table below lists supported environment variables, their type, default value, and a short description.

| Environment variable | Type | Default | Description |
|---|---:|---:|---|
| `LOGLEVEL` | string | `info` | Logging level: `debug`, `info`, `warn`, `error` |
| `PROXY_LISTENPORT` | int | `8080` | Port the proxy listens on |
| `PROXY_SHUTDOWNTIMEOUT` | duration | `10s` | Shutdown timeout (Go duration, e.g., `10s`) |
| `CACHE_TTL` | duration | `30s` | Time-to-live for cached records |
| `CACHE_MAXSIZE` | int (bytes) | `1048576` | Total cache capacity in bytes (1 MB) |
| `CACHE_MAXRECORDSIZE` | int (bytes) | `1024` | Maximum allowed size per cached record in bytes |
| `PROXY_TARGETURL` | string | `http://httpbin.org` | Upstream target URL used by the proxy |

