# =====================================================================
# STAGE 1: Build the binary
# =====================================================================
FROM golang:1.25-alpine AS builder

# Install git and certificates (Alpine needs these for external packages/HTTPS)
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first to leverage Docker caching
COPY go.work go.work.sum ./
COPY apps/api/go.mod apps/api/go.sum ./apps/api/
RUN go work sync

# Copy the rest of the source code
COPY apps/api/ ./apps/api/

# Build the binary
# --mount=type=cache,target=/root/.cache/go-build allows caching of Go build artifacts between builds, speeding up subsequent builds
# - CGO_ENABLED=0 disables dynamic linking for a fully static binary
# - GOOS=linux ensures it targets the Linux container environment
# - ldflags="-s -w" strips debug information to shrink the file size
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./apps/api/main.go

# =====================================================================
# STAGE 2: Create the final lean image
# =====================================================================
FROM alpine:latest AS runner

# Add a non-root user for security
RUN adduser -D -g '' appuser

WORKDIR /app
    
# Copy the compiled binary and SSL certs from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/api /bin/api

# Use the non-root user
USER appuser

# Expose the port your app listens on
EXPOSE 8080

# Run the binary
CMD ["/bin/api"]
