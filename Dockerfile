FROM golang:tip-alpine AS builder

RUN apk --no-cache add git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app/arian-parser ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add curl ca-certificates tzdata && \
    adduser -D -s /bin/sh arian

# create directories for certificates
RUN mkdir -p /certs && chown arian:arian /certs

COPY --from=builder /app/arian-parser /app/arian-parser
RUN chmod +x /app/arian-parser

# switch to non-root user
USER arian

# expose standard SMTP ports
EXPOSE 25 587 2525

# health check endpoint
EXPOSE 8080

ENTRYPOINT ["/app/arian-parser"]