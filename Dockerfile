FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o arian ./cmd/

FROM gcr.io/distroless/static-debian11:latest

WORKDIR /app
COPY --from=builder /app/arian /app/arian
ENTRYPOINT ["/app/arian"]