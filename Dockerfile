# ----- build ----------

FROM golang:1.24.6-alpine AS builder

ARG BUILD_TIME
ARG GIT_COMMIT
ARG GIT_BRANCH

# install dependencies
RUN apk add --no-cache git ca-certificates curl make

WORKDIR /build

# copy dependencies
COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

# copy application source
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# build the binary with optimizations
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath -ldflags="-s -w \
        -X 'arian-parser/internal/version.BuildTime=${BUILD_TIME:-$(date -u +%Y%m%d-%H%M%S)}' \
        -X 'arian-parser/internal/version.GitCommit=${GIT_COMMIT:-dev}' \
        -X 'arian-parser/internal/version.GitBranch=${GIT_BRANCH:-main}' \
        -extldflags '-static'" \
    -a -installsuffix cgo \
    -o arian-parser ./cmd/main.go

# download grpc_health_probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.39 && \
    curl -sL "https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64" \
    -o /build/grpc_health_probe && \
    chmod +x /build/grpc_health_probe

# ----- runtime ----------
FROM alpine:3.22.1

# metadata labels
LABEL org.opencontainers.image.title="arian-email-parser" \
      org.opencontainers.image.description="SMTP to gRPC Email Parser" \
      org.opencontainers.image.vendor="Arian" \
      org.opencontainers.image.source="https://github.com/xhos/arian-email-parser"

# install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    openssl && \
    # create non-root user
    addgroup -g 1001 -S arian && \
    adduser -u 1001 -S -G arian -h /app -s /bin/false arian && \
    # create certificate directory
    mkdir -p /certs /app && \
    chown -R arian:arian /certs /app

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binaries from builder
COPY --from=builder --chown=arian:arian /build/arian-parser /app/arian-parser
COPY --from=builder --chown=arian:arian /build/grpc_health_probe /usr/local/bin/grpc_health_probe

# Set environment defaults
ENV TZ=UTC \
    SMTP_ADDR=:25 \
    EMAIL_PARSER_GRPC_PORT=50053 \
    SMTP_DOMAIN=localhost \
    LOG_LEVEL=info

# Certificate volume for TLS
VOLUME ["/certs"]

# Switch to non-root user
USER arian
WORKDIR /app

# Expose ports
# 25: SMTP with STARTTLS
# 50053: gRPC health/status
EXPOSE 25 50053

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD ["/usr/local/bin/grpc_health_probe", "-addr=:50053"]

# Run the service
ENTRYPOINT ["/app/arian-parser"]