# List available commands
default:
    @just --list

# Generate Go code from templ files once
generate:
    templ generate

# Live reload for development with FLAGs
dev +args="":
    wgo -file=.go -file=.templ -xfile=_templ.go \
    templ generate ./ui/... :: go run ./cmd/web {{args}}

# Run tests with file watching
watch-test:
    wgo go test -v ./...

# Build the production binary
build: generate
    go build -o bin/snippetbox ./cmd/web

# Clean up build artifacts
clean:
    rm -rf bin/
    fd "_templ.go$" -x rm
