# Build stage
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files and download dependencies
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tracker ./cmd/tracker

# Final stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the compiled binary
COPY --from=builder /app/tracker .

# Create a non-root user
RUN adduser -D -g '' appuser
USER appuser

# Create config directory
RUN mkdir -p /app/config

# Set environment variables
ENV SOLANA_RPC_ENDPOINT=https://api.mainnet-beta.solana.com
ENV SOLANA_WS_ENDPOINT=wss://api.mainnet-beta.solana.com

# Command to run
ENTRYPOINT ["/app/tracker"]