#!/bin/bash

echo "🧪 Running MCP Registry Test Suite"
echo "=================================="

# Set test environment
export MCP_ENVIRONMENT=test
export MCP_LOG_LEVEL=error

# Run all tests with verbose output
echo "📋 Running unit tests..."
go test -v ./internal/...

# Check exit code
if [ $? -eq 0 ]; then
    echo ""
    echo "✅ All tests passed!"
else
    echo ""
    echo "❌ Some tests failed!"
    exit 1
fi

echo ""
echo "🏃‍♂️ Running race condition tests..."
go test -race ./internal/...

if [ $? -eq 0 ]; then
    echo "✅ No race conditions detected!"
else
    echo "❌ Race conditions detected!"
    exit 1
fi