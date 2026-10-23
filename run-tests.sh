#!/usr/bin/env bash
# Builds the test image, runs the tests, and exits with the tests' status code.
set -uo pipefail
cd "$(dirname "$0")"

IMAGE="go-rest-api-test:latest"

if ! docker build -f Dockerfile.test -t "$IMAGE" .; then
  echo "❌ Failed to build test image"
  exit 1
fi

docker run --rm "$IMAGE"
STATUS=$?

if [ "$STATUS" -eq 0 ]; then
  echo "✅ All tests passed"
else
  echo "❌ Tests failed (exit code $STATUS)"
fi
exit "$STATUS"