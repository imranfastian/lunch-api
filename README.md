# Lunch API

## Background

SciLifeLab has offices in Stockholm and Uppsala. Every day, employees need to decide where to go for lunch. This API serves the 15 local restaurants (with their weekly menus) from a JSON data file, and provides several endpoints to help employees make that decision:

- Browse all restaurants, filtered by city
- See **today's menu** without having to look through the full weekly schedule
- Find restaurants **near the SciLifeLab office** in your city, sorted by walking distance

---

## API endpoints

| Method | Path                              | Description                                                                                                                              |
| ------ | --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `GET`  | `/api/restaurants`                | List all restaurants. `?city=stockholm\|uppsala` filters by city (case-insensitive).                                                     |
| `GET`  | `/api/restaurants/today`          | All restaurants with **only today's menu items**. Supports `?city=`.                                                                     |
| `GET`  | `/api/restaurants/nearby`         | Restaurants within `?radius=` km of the SciLifeLab office. `?city=` is required; default radius is 5 km. Results are sorted by distance. |
| `GET`  | `/api/restaurants/:id`            | Details for a single restaurant.                                                                                                         |
| `GET`  | `/api/restaurants/:id/menu`       | Full weekly menu for a restaurant.                                                                                                       |
| `GET`  | `/api/restaurants/:id/menu/today` | Only today's menu items for a restaurant.                                                                                                |

### Example requests

```bash
# All restaurants in Uppsala
curl http://localhost:8000/api/restaurants?city=Uppsala

# Today's lunch options in Stockholm
curl http://localhost:8000/api/restaurants/today?city=stockholm

# Restaurants within 3 km of SciLifeLab Uppsala (Husargatan 3, BMC)
curl "http://localhost:8000/api/restaurants/nearby?city=uppsala&radius=3"

# Today's menu for restaurant 1
curl http://localhost:8000/api/restaurants/1/menu/today
```

### Distance / nearby logic

The `/nearby` endpoint uses the **Haversine great-circle formula** against hardcoded SciLifeLab office coordinates:

| City      | Office                             |
| --------- | ---------------------------------- |
| Stockholm | Tomtebodavägen 23, 171 65 Solna    |
| Uppsala   | Husargatan 3 (BMC), 751 23 Uppsala |

Approximate GPS coordinates for each restaurant are kept in [`src/handlers/geo.go`](src/handlers/geo.go) alongside the formula, so the provided `restaurants_menu.json` data file remains unchanged.

---

## Project structure

```
.
├── src/
│   ├── main.go              # Entry point: config load, router setup, metrics server on :9091
│   ├── middleware/
│   │   ├── metrics.go       # Prometheus middleware — HTTP request counter + latency histogram
│   │   └── logger.go        # Structured request logger
│   ├── config/
│   │   └── config.go        # Data types and in-memory config loader (reads JSON files)
│   ├── handlers/
│   │   ├── restaurant.go    # All restaurant HTTP handlers
│   │   └── geo.go           # Haversine formula + SciLifeLab office & restaurant coordinates
│   ├── routes/
│   │   └── routes.go        # Central route registration
│   └── data/
│       └── restaurants_menu.json   # Provided seed data (15 restaurants)
├── docs/
│   ├── docs.go              # OpenAPI spec embedded as Go string (registered via init())
│   ├── swagger.json         # Same spec as a static JSON file (for tooling / reference)
│   └── deployment.md        # Step-by-step deployment guide (kind, kustomize, helm)
├── tests/
│   ├── restaurant_test.go   # Handler integration tests (19 tests)
│   └── config_test.go       # Config loader test
├── k8s/
│   ├── base/                # Core manifests (namespace, configmap, deployment, service, ingress, hpa)
│   ├── overlays/
│   │   ├── dev/             # Kustomize overlay — 1 replica, debug mode
│   │   └── prod/            # Kustomize overlay — 3 replicas, pinned image tag
│   ├── helm/lunch-api/      # Helm chart (alternative to kustomize)
│   ├── monitoring/          # Prometheus + Grafana plain Kubernetes manifests
│   │   ├── prometheus/
│   │   │   ├── configmap.yaml   # Prometheus scrape config (pod auto-discovery via annotations)
│   │   │   ├── deployment.yaml  # Prometheus pod with config + persistent data volumes
│   │   │   ├── service.yaml     # ClusterIP on port 9090
│   │   │   ├── pvc.yaml         # 1 Gi PersistentVolumeClaim for /prometheus data directory
│   │   │   └── rbac.yaml        # ServiceAccount + ClusterRole to query the Kubernetes API
│   │   ├── grafana/
│   │   │   ├── configmap.yaml   # Datasource (Prometheus URL), dashboard provider, dashboard JSON
│   │   │   ├── deployment.yaml  # Grafana pod with provisioning volumes mounted
│   │   │   └── service.yaml     # ClusterIP on port 3000
│   │   └── kustomization.yaml   # Applies the full monitoring stack in one command
│   └── kind/cluster.yaml    # kind cluster config (1 control-plane + 2 workers, k8s v1.30.0)
├── prometheus/
│   └── prometheus.yml       # Prometheus config for local docker-compose (scrapes host :9091)
├── grafana/
│   ├── provisioning/        # Auto-loaded datasource + dashboard provider config
│   └── dashboards/          # Lunch API dashboard JSON
├── docker-compose.yml       # Local observability stack: Prometheus :9090 + Grafana :3000
└── Dockerfile               # Multi-stage build → distroless runtime image
```

---

## Running locally

**Prerequisites:** Go 1.25+ — [go.dev/dl](https://go.dev/dl/)

```bash
# Clone and enter the repository
git clone <repo-url>
cd lunch-api

# Run the server (default port 8000, metrics on port 9091)
# Go downloads all dependencies automatically on first run
go run ./src/main.go
```

The server starts at `http://localhost:8000`.  
Interactive Swagger UI: `http://localhost:8000/swagger/index.html`  
Prometheus metrics: `http://localhost:9091/metrics`

---

## Running tests

```bash
go test ./...
```

Tests use `net/http/httptest` — no running server required. The test suite covers all endpoints including filtering, proximity search, and error cases.

```bash
# With verbose output
go test -v ./tests/
```

---

## Observability

The API exposes Prometheus metrics on a **separate port** (`:9091`) so scrape traffic never inflates the application metrics.

### Metrics exposed

| Metric | Type | Description |
|--------|------|-------------|
| `lunch_api_http_requests_total` | Counter | Total requests, labelled by `path`, `method`, `status` |
| `lunch_api_http_duration_seconds` | Histogram | Request latency with `path` and `method` labels |

### Local observability stack (docker-compose)

Runs Prometheus + Grafana as Docker containers against the locally running Go app.

```bash
# 1. Start the Go app (metrics server on :9091)
go run ./src/main.go

# 2. Start Prometheus + Grafana
docker compose -f docker-compose.yml up -d

# 3. Open Prometheus:  http://localhost:9090
#    Open Grafana:     http://localhost:3000  (admin / admin)
```

Prometheus scrapes `host.docker.internal:9091/metrics` every 15 s. Grafana auto-loads the Prometheus datasource and the Lunch API dashboard from the `grafana/` directory — no manual setup needed.

### In-cluster observability (kind)

Prometheus and Grafana run as pods inside the `lunch-api` namespace alongside the application. Prometheus discovers pods automatically via annotations — any pod annotated with `prometheus.io/scrape: "true"` and `prometheus.io/port` is scraped without changing the Prometheus config.

```bash
# Deploy the monitoring stack
kubectl apply -k k8s/monitoring/

# Access Grafana (keep this terminal open)
kubectl port-forward -n lunch-api svc/grafana 3000:3000

# Access Prometheus UI (optional, separate terminal)
kubectl port-forward -n lunch-api svc/prometheus 9090:9090
```

Open `http://localhost:3000` → login `admin / admin` → the **Lunch API** dashboard is pre-loaded.

#### Why RBAC is needed for in-cluster Prometheus

Prometheus uses the Kubernetes API to discover pods (`kubernetes_sd_configs`). The `rbac.yaml` manifest grants it a ServiceAccount with `get/list/watch` on pods, services, and endpoints cluster-wide. Without it, Prometheus cannot query the API and finds no targets.

#### Namespace

Everything — the application and the monitoring stack — lives in the single `lunch-api` namespace. Grafana reaches Prometheus via the short service DNS name `http://prometheus:9090` (no cross-namespace DNS needed).

---

## Kubernetes deployment

The application is packaged as a Docker image and deployed to Kubernetes via Kustomize overlays. A Helm chart is provided as an alternative install path. The GitHub Actions pipeline automates testing, building, and deployment on every push.

### File layout

```
k8s/
├── base/                     # Shared manifests applied to every environment
│   ├── namespace.yaml        #   Namespace: lunch-api
│   ├── configmap.yaml        #   GIN_MODE, PORT
│   ├── deployment.yaml       #   Security-hardened pods, probes, resource limits, metrics port 9091
│   ├── service.yaml          #   ClusterIP — port 80 (HTTP) + 9091 (metrics)
│   ├── ingress.yaml          #   nginx Ingress for lunch-api.local
│   └── hpa.yaml              #   Autoscaling: CPU + memory
├── overlays/
│   ├── dev/                  # 1 replica, GIN_MODE=debug, local image tag
│   └── prod/                 # 3 replicas, GIN_MODE=release, pinned to v1.0.0
├── helm/lunch-api/           # Helm chart — alternative to kustomize
├── monitoring/               # Prometheus + Grafana — same lunch-api namespace
│   ├── prometheus/           #   ConfigMap, Deployment, Service, PVC (1 Gi), RBAC
│   ├── grafana/              #   ConfigMap (datasource + dashboard), Deployment, Service
│   └── kustomization.yaml    #   kubectl apply -k k8s/monitoring/ deploys the full stack
└── kind/cluster.yaml         # kind cluster config (1 control-plane + 2 workers, k8s v1.30.0)
```

The CI/CD pipeline lives at [`.github/workflows/deploy.yaml`](.github/workflows/deploy.yaml) — GitHub Actions reads only from `.github/workflows/`.

### Complete CI/CD flow

Every push to `develop` runs tests, builds a Docker image, pushes it to GHCR with the `:develop` tag, then deploys to a short-lived kind cluster on the runner and smoke-tests three endpoints. Pushing a semver tag (`v1.0.0`) additionally produces `:v1.0.0`, `:v1.0`, and `:latest` image tags and triggers the prod deploy job (requires `KUBECONFIG_PROD` secret and manual approval via GitHub Environments). PRs only run tests — no image is pushed until the PR is merged.

### Where Helm fits

Kustomize and Helm both parameterise Kubernetes YAML for different environments but work differently:

|           | Kustomize                                  | Helm                                                         |
| --------- | ------------------------------------------ | ------------------------------------------------------------ |
| Approach  | Patches layered on top of plain YAML       | Template engine with a `values.yaml` file                    |
| Best for  | Your own cluster — precise overlay control | Distributing the app so others can install it in one command |
| Used here | CI/CD pipeline (`kubectl apply -k`)        | Demo install (`helm install`)                                |

**Rule of thumb:** use kustomize in CI/CD for day-to-day deployments; hand someone a Helm chart when they need to install without understanding the internals.

```bash
# Deploy in one command:
helm install lunch-api ./k8s/helm/lunch-api \
  --namespace lunch-api --create-namespace \
  --set image.tag=v1.0.0

# Upgrade to the next release:
helm upgrade lunch-api ./k8s/helm/lunch-api \
  --namespace lunch-api --set image.tag=v1.1.0
```

### Credentials

| Credential        | What it is                                     | How it arrives                                        |
| ----------------- | ---------------------------------------------- | ----------------------------------------------------- |
| `GITHUB_TOKEN`    | Short-lived token for pushing images to GHCR   | GitHub injects it per-run — you configure nothing     |
| `KUBECONFIG_PROD` | Base64-encoded kubeconfig for a remote cluster | Add once in **GitHub → Settings → Secrets → Actions** |

The image is built from a two-stage [Dockerfile](Dockerfile):

- **Stage 1** — Go 1.25 Alpine builder, `CGO_ENABLED=0`
- **Stage 2** — `distroless/static:nonroot` (no shell, UID 65532, read-only filesystem)
