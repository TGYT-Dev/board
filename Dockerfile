# Build stage
FROM golang:alpine AS builder

WORKDIR /src

# Copy go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/board .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/board /app/board

# Create data directory for database and SSH host key persistence
RUN mkdir -p /app/data /app/.ssh

ENV PORT=2222 \
    HOST=0.0.0.0 \
    DB_PATH=/app/data/board.db \
    KEY_PATH=/app/data/.ssh/id_ed25519

EXPOSE 2222

# VOLUME ["/app/data"]

ENTRYPOINT ["/app/board"]
