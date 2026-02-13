# ---------- BUILD STAGE ----------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Source
COPY . .

# Build all services
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o notification-api ./cmd/notification-api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o outbox-publisher ./cmd/outbox-publisher
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sms-worker ./cmd/sms-worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o email-worker ./cmd/email-worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o push-worker ./cmd/push-worker

# ---------- RUNTIME STAGE ----------
FROM alpine:3.19

WORKDIR /app

# SSL certificates (Kafka + HTTPS için önemli)
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/notification-api .
COPY --from=builder /app/outbox-publisher .
COPY --from=builder /app/sms-worker .
COPY --from=builder /app/email-worker .
COPY --from=builder /app/push-worker .

EXPOSE 8080

# Default (compose override edecek)
CMD ["./notification-api"]
