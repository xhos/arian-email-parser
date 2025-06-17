FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o arian-parser ./cmd/

FROM gcr.io/distroless/base-debian11:latest

WORKDIR /app
COPY --from=curlimages/curl:latest /usr/bin/curl /usr/bin/curl
COPY --from=builder /app/arian-parser /app/arian-parser
ENTRYPOINT ["/app/arian-parser"]