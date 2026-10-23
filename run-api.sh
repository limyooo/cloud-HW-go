#!/usr/bin/env bash
# Builds the API image and runs it until stopped (Ctrl+C or `docker stop go-rest-api`).
set -euo pipefail
cd "$(dirname "$0")"

IMAGE="go-rest-api:latest"
PORT="${PORT:-8080}"

docker build -f Dockerfile.api -t "$IMAGE" .

# Attach a TTY only when run from an interactive terminal, so Ctrl+C works locally
# and the script still runs in non-interactive environments.
TTY_FLAGS=""
if [ -t 0 ] && [ -t 1 ]; then
  TTY_FLAGS="-it"
fi

echo "Starting API on http://localhost:${PORT}  (press Ctrl+C to stop)"
# shellcheck disable=SC2086
exec docker run --rm $TTY_FLAGS --name go-rest-api -p "${PORT}:8080" "$IMAGE"