# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install templ CLI and generate Go code
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001
RUN /go/bin/templ generate ./ui/...

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o snippetbox ./cmd/web

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS (if needed)
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/snippetbox .

# Expose port
EXPOSE 4000

# Run the application
CMD ["./snippetbox"]
