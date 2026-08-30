# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Install templ CLI (cached unless go.mod/go.sum changes)
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001

# Copy the rest of the source code
COPY . .

# Generate Go code from templ files
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
