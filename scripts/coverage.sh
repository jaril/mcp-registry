#!/bin/bash

echo "📊 Generating Test Coverage Report"
echo "================================="

# Create coverage directory
mkdir -p coverage

# Run tests with coverage
echo "Running tests with coverage..."
go test -coverprofile=coverage/coverage.out ./internal/...

# Generate HTML report
echo "Generating HTML coverage report..."
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# Show coverage summary
echo ""
echo "📋 Coverage Summary:"
go tool cover -func=coverage/coverage.out

echo ""
echo "📂 HTML report generated: coverage/coverage.html"
echo "💡 Open in browser: open coverage/coverage.html"