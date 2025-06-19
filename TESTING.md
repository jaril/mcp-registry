# Run all tests

`go test ./internal/...`

# Run tests with verbose output

`go test -v ./internal/...`

# Run tests with race detection

`go test -race ./internal/...`

# Use the test script

`./scripts/test.sh`

# Generate coverage report

`./scripts/coverage.sh`

# Open coverage report in browser

`open coverage/coverage.html`

# Run only model tests

`go test ./internal/models/`

# Run only handler tests

`go test ./internal/handlers/`

# Run specific test

`go test -run TestMemoryStore_Create ./internal/storage/`

# Run benchmarks

`go test -bench=. ./internal/storage/`
