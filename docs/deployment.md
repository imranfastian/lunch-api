# Deployment Guide — Lunch API

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Docker Desktop | ≥ 4.x | [docs.docker.com](https://docs.docker.com/desktop/) |
| kind | ≥ 0.20 | `go install sigs.k8s.io/kind@latest` |
| kubectl | ≥ 1.28 | Bundled with Docker Desktop |
| kustomize | ≥ 5.x | `go install sigs.k8s.io/kustomize/kustomize/v5@latest` |
| helm | ≥ 3.x | [helm.sh/docs/intro/install](https://helm.sh/docs/intro/install/) |

---

## 1. Build the Docker image

```bash
docker build -t lunch-api:dev .
```

Two-stage build: `golang:1.25-alpine` compiles the binary → `distroless/static:nonroot` runs it (no shell, UID 65532, ~8 MB image).

---

## 2. Create the kind cluster

The cluster config is at [`k8s/kind/cluster.yaml`](../k8s/kind/cluster.yaml) — 1 control-plane + 2 workers, pinned to k8s v1.30.0.

```bash
# Windows PowerShell
kind create cluster --name lunch --config k8s\kind\cluster.yaml --wait 90s

# Linux / macOS
kind create cluster --name lunch --config k8s/kind/cluster.yaml --wait 90s

# Delete when done
kind delete cluster --name lunch
```

---

## 3. Load the image into kind

kind nodes have their own isolated containerd — `kind load` copies the image in so kubelet can start pods without a registry pull.

```bash
kind load docker-image lunch-api:dev --name lunch
```

---

## 4. Deploy with Kustomize

```bash
# Dev overlay (1 replica, GIN_MODE=debug)
kubectl apply -k k8s/overlays/dev
kubectl rollout status deployment/lunch-api -n lunch-api

# Prod overlay (3 replicas, pinned image tag)
kubectl apply -k k8s/overlays/prod
kubectl rollout status deployment/lunch-api -n lunch-api
```

---

## 5. Deploy with Helm (alternative)

```bash
# Install
helm install lunch-api ./k8s/helm/lunch-api \
  --namespace lunch-api --create-namespace \
  --set image.tag=dev

# Upgrade to a new image tag
helm upgrade lunch-api ./k8s/helm/lunch-api \
  --namespace lunch-api --set image.tag=v1.0.0

# Uninstall
helm uninstall lunch-api --namespace lunch-api
```

---

## 6. Access the API

```bash
kubectl port-forward svc/lunch-api 8000:80 -n lunch-api
```

Keep the terminal open while browsing:

- API: `http://localhost:8000/api/restaurants`
- Swagger UI: `http://localhost:8000/swagger/index.html`

---

## 7. Verify the deployment

```bash
# Pods running across worker nodes
kubectl get pods -n lunch-api -o wide

# Application logs
kubectl logs -l app=lunch-api -n lunch-api -f

# Quick smoke test
curl http://localhost:8000/api/restaurants
curl "http://localhost:8000/api/restaurants/nearby?city=Uppsala&radius=2"
```

---

## Security

The container runs with the minimum required privileges:

- Non-root user (UID 65532, distroless nonroot)
- Read-only filesystem
- All Linux capabilities dropped
- No privilege escalation
- CPU and memory limits enforced per pod
- Pods spread across worker nodes via `topologySpreadConstraints`

---

## CI/CD

The pipeline at `.github/workflows/deploy.yaml` runs on every push to `develop` or a semver tag:

1. **Test** — `go test -v -race ./...`
2. **Build & push** — Docker image to GHCR (`:develop` or `:v1.0.0 :v1.0 :latest`)
3. **Smoke test** — kind cluster on the runner, `kubectl apply -k k8s/overlays/dev`, curl three endpoints
4. **Prod deploy** — `kubectl apply -k k8s/overlays/prod` on semver tags (requires `KUBECONFIG_PROD` secret and manual approval)
