# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/board .

# Create directories in builder so they exist in final image
RUN mkdir -p /app/data /app/.ssh

# Runtime stage: lighter than alpine, no package manager
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/board /app/board
# Copy empty directories created in builder so volume mount points exist
COPY --from=builder /app/data /app/data
COPY --from=builder /app/.ssh /app/.ssh

ENV PORT=2222 \
    HOST=0.0.0.0 \
    DB_PATH=/app/data/board.db \
    KEY_PATH=/app/data/.ssh/id_ed25519

EXPOSE 2222

VOLUME ["/app/data"]

ENTRYPOINT ["/app/board"]
