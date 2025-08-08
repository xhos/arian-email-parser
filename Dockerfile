FROM golang:tip-alpine AS builder

ARG VERSION=unknown
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
ARG GIT_BRANCH=unknown

RUN apk --no-cache add git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w \
    -X arian-parser/internal/version.BuildTime=${BUILD_TIME} \
    -X arian-parser/internal/version.GitCommit=${GIT_COMMIT} \
    -X arian-parser/internal/version.GitBranch=${GIT_BRANCH}" \
    -o /app/arian-parser ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add curl ca-certificates tzdata && \
    adduser -D -s /bin/sh arian

# Install grpc_health_probe
RUN apk --no-cache add curl ca-certificates && \
    curl -L https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.39/grpc_health_probe-linux-amd64 -o /usr/local/bin/grpc_health_probe && \
    chmod +x /usr/local/bin/grpc_health_probe

# create directories for certificates
RUN mkdir -p /certs && chown arian:arian /certs

COPY --from=builder /app/arian-parser /app/arian-parser
RUN chmod +x /app/arian-parser

# switch to non-root user
USER arian

EXPOSE 25 50053

ENTRYPOINT ["/app/arian-parser"]
