# URL Shortener — Cloud-Native with Full Observability

[![CI](https://github.com/likhi/url-shortener/actions/workflows/ci.yml/badge.svg)](https://github.com/likhi/url-shortener/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A production-grade URL shortener built with Go, deployed on Kubernetes, and instrumented end-to-end with Prometheus and Grafana. Built to demonstrate backend engineering, SRE, and DevOps best practices.

---

## Architecture

```
                        ┌──────────────────────────────┐
                        │         Internet / DNS        │
                        └──────────────┬───────────────┘
                                       │ HTTPS
                        ┌──────────────▼───────────────┐
                        │   Ingress (nginx + TLS)       │
                        │   cert-manager auto-renew     │
                        └──────┬───────────────┬────────┘
                               │               │
               ┌───────────────▼──┐     ┌──────▼───────────────┐
               │  url-shortener   │ ... │   url-shortener       │
               │  pod (Go/Gin)    │     │   pod (Go/Gin)        │
               │  ─────────────   │     │   ─────────────       │
               │  • POST /urls    │     │   • POST /urls        │
               │  • GET /:code    │     │   • GET /:code        │
               │  • GET /metrics  │     │   • GET /metrics      │
               └──────┬────┬──────┘     └──────┬────┬──────────┘
                      │    │                   │    │
          ┌───────────┘    └──────┐  ┌─────────┘    └──────────┐
          │                      │  │                          │
 ┌────────▼────────┐    ┌────────▼──▼────────┐                │
 │   PostgreSQL    │    │       Redis         │                │
 │  (primary store)│    │ cache + rate-limit  │                │
 └─────────────────┘    └────────────────────┘                │
                                                               │
          ┌────────────────────────────────────────────────────┘
          │              Observability stack
          │
 ┌────────▼────────┐    ┌─────────────────────┐
 │   Prometheus    │───►│      Grafana         │
 │  scrapes /metrics    │  dashboards + alerts │
 └─────────────────┘    └─────────────────────┘

 HPA autoscales pods 3 → 20 replicas on CPU/Memory pressure.
 Background worker sweeps expired URLs every hour.
 Redis-backed rate limiter works across all replicas (100 req/min/IP).
```

### Request flow — redirect

```
Browser ──GET /abc1234──► pod
                           │
                    Redis hit? ──yes──► 301 redirect (< 1ms)
                           │
                          no
                           │
                    Postgres query ──► cache ──► 301 redirect
                           │
                    async: increment click_count
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP framework | Gin |
| Database | PostgreSQL 16 (pgx/v5 driver) |
| Cache / Rate-limit | Redis 7 |
| Metrics | Prometheus client_golang |
| Dashboards | Grafana |
| Containerisation | Docker (distroless image) |
| Orchestration | Kubernetes + HPA |
| Ingress | nginx-ingress + cert-manager |
| CI/CD | GitHub Actions |
| Config | Viper (env-var driven) |
| Logging | Zap (structured JSON) |

---

## Quick Start — run locally in 3 commands

**Prerequisites:** Docker Desktop (includes `docker compose`)

```bash
git clone https://github.com/likhi/url-shortener.git
cd url-shortener

cp .env.example .env          # edit values if needed (defaults work out of the box)

docker compose -f deployments/docker/docker-compose.yml up --build
```

Services started:

| Service | URL |
|---|---|
| API | http://localhost:8080 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3000 (admin / admin) |

The database schema is applied automatically on first boot.

---

## API Reference

### Shorten a URL

```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"original_url": "https://example.com"}'
```

```json
{
  "short_code": "aB3xK7q",
  "short_url": "http://localhost:8080/aB3xK7q",
  "original_url": "https://example.com",
  "created_at": "2026-05-20T10:00:00Z"
}
```

### Shorten with a custom code and expiry

```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://example.com/very/long/path",
    "custom_code": "mylink",
    "expires_at": "2026-12-31T23:59:59Z"
  }'
```

### Redirect

```bash
curl -L http://localhost:8080/aB3xK7q
# → 301 to https://example.com
```

### Get click statistics

```bash
curl http://localhost:8080/api/v1/urls/aB3xK7q/stats
```

```json
{
  "short_code": "aB3xK7q",
  "original_url": "https://example.com",
  "click_count": 42,
  "created_at": "2026-05-20T10:00:00Z"
}
```

### Health check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### HTTP status codes

| Code | Meaning |
|---|---|
| 201 | URL created |
| 301 | Redirect |
| 400 | Invalid request body / URL |
| 404 | Short code not found or expired |
| 409 | Custom code already taken |
| 429 | Rate limit exceeded (100 req/min/IP) |
| 500 | Internal server error |

---

## Project Structure

```
.
├── cmd/server/main.go              # entrypoint — wires everything, graceful shutdown
├── internal/
│   ├── api/
│   │   ├── handlers/               # Gin handlers (Create, Redirect, Stats, Health)
│   │   ├── middleware/             # CORS, Zap logger, Prometheus, rate limiter
│   │   └── routes.go              # route registration
│   ├── config/config.go            # env-var config via Viper
│   ├── database/
│   │   ├── cache.go               # Cache interface + RedisCache adapter
│   │   ├── migrate.go             # embedded-SQL migration runner
│   │   ├── postgres.go            # pgxpool connection
│   │   └── redis.go               # Redis client
│   ├── models/url.go              # domain types + request/response structs
│   ├── repository/                # PostgreSQL data access + URLStore interface
│   ├── service/                   # business logic + URLShortener interface
│   └── worker/cleanup.go          # background job: delete expired URLs hourly
├── pkg/
│   ├── metrics/prometheus.go      # all Prometheus counters / histograms / gauges
│   └── shortener/generator.go     # crypto-random short code generator
├── migrations/                    # *.up.sql / *.down.sql — embedded into binary
├── deployments/
│   ├── docker/                    # Dockerfile + docker-compose.yml
│   └── k8s/                       # namespace, configmap, secret, deployment,
│                                  # service, ingress, hpa
├── monitoring/
│   ├── prometheus/prometheus.yml  # scrape config
│   └── grafana/                   # auto-provisioned datasource + dashboard
└── .github/workflows/             # CI (test + lint + build) + CD (push + deploy)
```

---

## Development

```bash
# Run all tests (with race detector)
go test ./... -race

# Build binary
make build          # output: bin/url-shortener

# Run locally (needs Postgres + Redis running)
make run

# Spin up full stack
make docker-up

# Tear down
make docker-down

# Run linter (requires golangci-lint)
make lint
```

### Environment variables

All config is driven by env vars (see `.env.example`):

| Variable | Default | Description |
|---|---|---|
| `SERVER_PORT` | `8080` | HTTP listen port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_USER` | — | PostgreSQL user |
| `DB_PASSWORD` | — | PostgreSQL password |
| `DB_NAME` | — | PostgreSQL database |
| `REDIS_HOST` | `localhost` | Redis host |
| `APP_BASE_URL` | `http://localhost:8080` | Prefix for generated short URLs |
| `APP_SHORT_CODE_LEN` | `7` | Length of auto-generated codes |
| `APP_ENV` | `development` | `development` or `production` |

---

## Observability

### Prometheus metrics

| Metric | Type | Description |
|---|---|---|
| `url_shortener_urls_created_total` | Counter | Total short URLs created |
| `url_shortener_redirects_total{short_code}` | Counter | Redirects per code |
| `url_shortener_cache_hits_total` | Counter | Redis cache hits |
| `url_shortener_cache_misses_total` | Counter | Redis cache misses |
| `url_shortener_http_request_duration_seconds` | Histogram | Request latency by method/path/status |
| `url_shortener_active_connections` | Gauge | Active HTTP connections |

### Grafana dashboard

The dashboard is auto-provisioned at startup. Open **http://localhost:3000** (admin / admin) and navigate to **URL Shortener**. It shows:

- URLs created (total counter)
- Cache hit rate (gauge, 0–1)
- Redirects per second (time-series, per short code)
- HTTP latency p99 (time-series, per route)
- Active connections (stat)

---

## Kubernetes Deployment

```bash
# Apply all manifests
kubectl apply -f deployments/k8s/

# Check rollout
kubectl rollout status deployment/url-shortener -n url-shortener

# Scale manually
kubectl scale deployment/url-shortener --replicas=5 -n url-shortener
```

The HPA automatically scales between **3 and 20 replicas** based on CPU (>60%) and memory (>70%) utilisation.

Update `deployments/k8s/secret.yaml` with real credentials before applying — or use your cloud provider's secrets manager and reference the secret from there.

---

## CI/CD

Every pull request triggers:

1. `go vet` + `golangci-lint`
2. `go test ./... -race`
3. Docker image build (validates the Dockerfile compiles)

Every push to `main` additionally:

4. Pushes the Docker image to GitHub Container Registry (`ghcr.io`)
5. Applies the Kubernetes manifests to the target cluster (rolling update)

See [`.github/workflows/`](.github/workflows/) for the full pipeline definitions.

---

## License

MIT
