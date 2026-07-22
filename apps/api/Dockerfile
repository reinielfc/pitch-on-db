# =====================================================================
# STAGE 1: Build the binary
# =====================================================================
FROM golang:1.26-alpine@sha256:9097beb5536220f7857bdcb65c1b4b340630dd7a70b85f03d5af29640b06693d AS builder

# Install git and certificates (Alpine needs these for external packages/HTTPS)
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first to leverage Docker caching
COPY go.mod go.sum ./

# Pre-download modules into a persistent cache so source-only changes do not re-fetch deps.
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source code
COPY . .

# Set the application version from build args (default to "dev" if not provided)
ARG APP_VERSION=dev

# Build the binary
# --mount=type=cache,target=/root/.cache/go-build caches compiled package artifacts between builds
# --mount=type=cache,target=/go/pkg/mod caches downloaded Go modules between builds
# - CGO_ENABLED=0 disables dynamic linking for a fully static binary
# - GOOS=linux ensures it targets the Linux container environment
# - ldflags="-s -w" strips debug information to shrink the file size
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux \
        go build \
            -ldflags="-s -w -X main.version=${APP_VERSION}" \
            -o /bin/pitchondb ./cmd/service

# =====================================================================
# STAGE 2: Create the final lean image
# =====================================================================
FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce AS runner

# Add a non-root user for security
RUN adduser -D -g '' appuser

WORKDIR /app
    
# Copy the compiled binary and SSL certs from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/pitchondb /bin/pitchondb

# Use the non-root user
USER appuser

# Expose the port your app listens on
EXPOSE 8080

# Run the binary
CMD ["/bin/pitchondb"]
