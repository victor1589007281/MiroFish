FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/swarm-predict ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/swarm-predict .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

ENV GIN_MODE=release

ENTRYPOINT ["./swarm-predict"]
