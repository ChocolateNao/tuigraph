# Default: show available recipes
default:
    @just --list

# Build all packages
build:
    go build ./...

# Run all tests
test:
    go test ./...

# Run tests with race detector
test-race:
    go test -race ./...

# Static analysis
vet:
    go vet ./...

# Run golangci-lint
lint:
    golangci-lint run ./...

# Run golangci-lint and fix all fixable errors
lint-fix:
    golangci-lint run ./... --fix

# Format check (must be clean before commit)
fmt-check:
    @gofmt_out=$$(gofmt -l .); \
    if [ -n "$$gofmt_out" ]; then \
        echo "Files need formatting:"; \
        echo "$$gofmt_out"; \
        exit 1; \
    fi

# Auto-format all source files
fmt:
    gofumpt -w .

# Run all checks: build, vet, lint, tests
check: build vet lint test fmt-check

# Clean build artifacts
clean:
    rm -rf tmp/

# Run the static demo (all diagram types)
demo:
    go run ./_example/07_static

# Run the live real-time demo
demo-live:
    go run ./_example/06_live

# Run demo without colors
demo-no-color:
    NO_COLOR=1 go run ./_example/07_static

# Run demo in pure ASCII mode
demo-ascii:
    LC_ALL=C NO_COLOR=1 go run ./_example/07_static

# Install golangci-lint (if not present)
setup-lint:
    @which golangci-lint > /dev/null 2>&1 || \
        go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Install gofumpt (if not present)
setup-fmt:
    @which gofumpt > /dev/null 2>&1 || \
        go install mvdan.cc/gofumpt@latest

# Install all dev tools
setup: setup-lint setup-fmt setup-lefthook

# Set up git hooks via lefthook (run once after clone)
setup-lefthook:
    @which lefthook > /dev/null 2>&1 || \
        go install github.com/evilmartians/lefthook/v2@latest

# Apply lefthook configuration
sync-hooks:
    lefthook install

# Run everything a CI would run
ci: check test-race
