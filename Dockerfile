# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files and download
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/im-here ./cmd/scanner

# Final stage
FROM alpine:latest

# Certificates are needed for HTTPS requests (GitHub API, Resend, Neon DB)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/bin/im-here .

# Run the scanner
CMD ["./im-here"]