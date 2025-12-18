# reverseProxyExample

## How to run

- Run the server locally:

```bash
make run
```

- Run the server with example environment variables:

```bash
LOGLEVEL=debug PROXY_SERVER_LISTENPORT=9090 CACHE_TTL=30s make run
```

## How to run tests

- Run the unit test suite:

```bash
make test
```

## Configuration (environment variables)

The table below lists supported environment variables, their type, default value, and a short description.

| Environment variable | Type | Default | Description |
|---|---:|---:|---|
| `LOGLEVEL` | string | `info` | Logging level: `debug`, `info`, `warn`, `error` |
| `PROXY_SERVER_LISTENPORT` | int | `8080` | Port the proxy listens on |
| `PROXY_SERVER_SHUTDOWNTIMEOUT` | duration | `10s` | Shutdown timeout (Go duration, e.g., `10s`) |
| `PROXY_SERVER_READTIMEOUT` | duration | `10s` | Server read timeout (Go duration, e.g., `10s`) |
| `PROXY_SERVER_WRITETIMEOUT` | duration | `10s` | Server write timeout (Go duration, e.g., `10s`) |
| `PROXY_SERVER_IDLETIMEOUT` | duration | `120s` | Server idle timeout (Go duration, e.g., `120s`) |
| `PROXY_TRANSPORT_MAXIDLECONNECTIONS` | int | `100` | Transport max idle connections |
| `PROXY_TRANSPORT_MAXIDLECONNSPERHOST` | int | `20` | Transport max idle connections per host |
| `PROXY_TRANSPORT_IDLECTIMEOUT` | duration | `90s` | Transport idle connection timeout |
| `PROXY_TRANSPORT_DIALTIMEOUT` | duration | `5s` | Transport dial timeout |
| `CACHE_TTL` | duration | `30s` | Time-to-live for cached records |
| `CACHE_MAXSIZE` | int (bytes) | `1048576` | Total cache capacity in bytes (1 MB) |
| `CACHE_MAXRECORDSIZE` | int (bytes) | `1024` | Maximum allowed size per cached record in bytes |
| `PROXY_TARGETURL` | string | `http://httpbin.org` | Upstream target URL used by the proxy |

