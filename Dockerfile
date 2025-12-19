# Build Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy gomod
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binaries
RUN go build -o producer ./cmd/producer/main.go
RUN go build -o worker ./cmd/worker/main.go

# Runtime Stage
FROM alpine:latest

WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/producer .
COPY --from=builder /app/worker .

# Default command (can be overridden)
CMD ["./producer"]
