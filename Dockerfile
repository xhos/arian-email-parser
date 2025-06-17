FROM golang:tip-alpine AS builder

RUN apk --no-cache add git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app/arian-parser ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add curl ca-certificates
COPY --from=builder /app/arian-parser /app/arian-parser
ENTRYPOINT ["/app/arian-parser"]