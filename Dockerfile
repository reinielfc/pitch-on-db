# =====================================================================
# STAGE 1: Build the binary
# =====================================================================
FROM golang:1.25-alpine AS builder

# Install git and certificates (Alpine needs these for external packages/HTTPS)
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
# - CGO_ENABLED=0 disables dynamic linking for a fully static binary
# - GOOS=linux ensures it targets the Linux container environment
# - ldflags="-s -w" strips debug information to shrink the file size
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/server

# =====================================================================
# STAGE 2: Create the final lean image
# =====================================================================
FROM alpine:latest AS runner

# Add a non-root user for security
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy the compiled binary and SSL certs from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main .

# Use the non-root user
USER appuser

# Expose the port your app listens on
ENV PORT=8080
EXPOSE ${PORT}

# Run the binary
CMD ["./main"]