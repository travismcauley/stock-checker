# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy the binary
COPY --from=builder /app/server /app/server

# Copy migrations
COPY --from=builder /app/migrations /app/migrations

# Expose port
EXPOSE 8080

# Run the server
CMD ["/app/server"]
