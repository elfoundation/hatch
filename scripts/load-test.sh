#!/usr/bin/env bash
set -euo pipefail

# Load test script for Hatch server.
# Assumes the server is already running at http://localhost:8080
# or at the URL specified by HATCH_URL.

HATCH_URL="${HATCH_URL:-http://localhost:8080}"
CONCURRENCY="${CONCURRENCY:-10}"
TOTAL="${TOTAL:-1000}"

echo "Running load test against $HATCH_URL"
echo "Concurrency: $CONCURRENCY, Total requests: $TOTAL"

# Build the load test binary if not present.
if [ ! -f ./loadtest ]; then
    echo "Building load test binary..."
    go build -o ./loadtest ./cmd/loadtest
fi

# Run the load test.
./loadtest -url "$HATCH_URL/healthz" -c "$CONCURRENCY" -n "$TOTAL"

echo "Load test completed."