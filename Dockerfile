# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache module downloads separately from source code
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a fully static binary (no CGO, no libc dependency)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOTOOLCHAIN=local \
    go build -ldflags="-w -s" -o server ./src/main.go

# ── Stage 2: Runtime ───────────────────────────────────────────────────────────
# distroless/static contains nothing except the CA certificates and tzdata.
# No shell, no package manager, minimal attack surface.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy the binary and the data file the server reads at startup
COPY --from=builder /app/server          ./server
COPY --from=builder /app/src/data        ./src/data

EXPOSE 8000

# nonroot (UID 65532) is provided by the distroless:nonroot image
USER nonroot:nonroot

ENTRYPOINT ["/app/server"]
