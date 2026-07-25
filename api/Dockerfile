# =====================================================================
# BASE STAGE: shared dependencies for both dev and builder stages
# =====================================================================
FROM golang:1.26-alpine3.23@sha256:622e56dbc11a8cfe87cafa2331e9a201877271cbff918af53d3be315f3da88cc AS base

# Install git (needed for `go get`/private modules) and CA certs (needed for HTTPS)
# Cache mount avoids re-downloading apk packages on every rebuild
RUN --mount=type=cache,target=/var/cache/apk \
    apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app

# Copy only dependency manifests first, so this layer (and go work sync below)
# stays cached as long as dependencies don't change, even if source code does
COPY go.work go.work.sum ./
COPY api/go.mod api/go.sum ./api/

# Sync the Go workspace and download modules
# Cache mount persists the module download cache across builds
RUN --mount=type=cache,target=/go/pkg/mod go work sync

# =====================================================================
# DEV STAGE: used by docker compose watch, includes hot-reload tooling
# =====================================================================
FROM base AS dev

# Install air for hot-reload on file changes during local development
RUN go install github.com/air-verse/air@latest

# Bring in the actual source (bind-mounted/synced over in practice via compose watch,
# but included here so `docker build --target dev` also works standalone)
COPY api/ ./api/

# Set working directory to the app source so air can find the .air.toml config file
WORKDIR /app/api

# Run air instead of the compiled binary — watches for changes and rebuilds/restarts
CMD ["air", "-c", ".air.toml"]

# =====================================================================
# BUILDER STAGE: compiles the static production binary
# =====================================================================
FROM base AS builder

# Copy full source now that deps are resolved and cached
COPY api/ ./api/

# Build a statically linked binary:
# - CGO_ENABLED=0 avoids dynamic linking so it runs on a minimal base (no libc dependency)
# - GOOS=linux targets the container's OS regardless of host OS
# - ldflags="-s -w" strips debug symbols/DWARF info to shrink binary size
# Cache mounts speed up recompiles by preserving build and module caches between builds
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./api/main.go

# =====================================================================
# RUNNER STAGE: minimal production image, no Go toolchain or source code
# =====================================================================
FROM alpine:3.22@sha256:fd791d74b68913cbb027c6546007b3f0d3bc45125f797758156952bc2d6daf40 AS runner

# Create a non-root user to run the app as (security best practice)
RUN adduser -D -g '' appuser

# Set working directory to /app, where the binary will be copied
WORKDIR /app

# Copy CA certs so the binary can make outbound HTTPS calls
# Copy the compiled binary — nothing else from the builder stage is needed
COPY --from=builder --chown=appuser:appuser /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder --chown=appuser:appuser /bin/api /bin/api

# Drop root privileges before running the app
USER appuser

# Document the port the app listens on (informational; doesn't actually publish it)
EXPOSE 8080

# Set Gin framework to release mode for production (disables debug output)
ENV GIN_MODE=release

# Default command for this stage — runs the compiled binary directly
CMD ["/bin/api"]
