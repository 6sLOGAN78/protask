# Multi-stage build for Go Backend
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the main executable and the cron runner
RUN CGO_ENABLED=0 GOOS=linux go build -o protask-server ./cmd/protask/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o protask-cron ./cmd/cron/main.go

# Start fresh with a small Alpine container
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates (needed for Clerk/Resend external APIs)
RUN apk --no-cache add ca-certificates

# Copy binaries from builder
COPY --from=builder /app/protask-server .
COPY --from=builder /app/protask-cron .
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

# Expose port
EXPOSE 8080

# Run entry point
CMD ["./protask-server"]
