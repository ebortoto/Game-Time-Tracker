#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [ -f ".env" ]; then
  set -a
  . "./.env"
  set +a
fi

SERVER_URL="${TRACKER_SERVER_URL:-http://localhost:8080}"
API_KEY="${TRACKER_API_KEY:-}"
CONFIG_PATH="${TRACKER_CLIENT_CONFIG:-config.json}"
OVERLAY="${TRACKER_OVERLAY:-true}"
START_HIDDEN="${TRACKER_START_HIDDEN:-true}"

CLIENT_BIN="./bin/tracker-client.exe"
if [ ! -f "$CLIENT_BIN" ]; then
  CLIENT_BIN="./bin/tracker-client"
fi

if [ ! -f "$CLIENT_BIN" ]; then
  echo "Client binary not found. Building ./cmd/client..."
  go build -o ./bin/tracker-client.exe ./cmd/client
  CLIENT_BIN="./bin/tracker-client.exe"
fi

ARGS=(
  "-config=$CONFIG_PATH"
  "-overlay=$OVERLAY"
  "-start-hidden=$START_HIDDEN"
  "-server-url=$SERVER_URL"
)

if [ -n "$API_KEY" ]; then
  ARGS+=("-api-key=$API_KEY")
fi

if [[ "${OSTYPE:-}" == msys* || "${OSTYPE:-}" == cygwin* || "${OSTYPE:-}" == win32* ]]; then
  if command -v powershell.exe >/dev/null 2>&1; then
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "./scripts/start-client.ps1"
    exit 0
  fi
fi

nohup "$CLIENT_BIN" "${ARGS[@]}" >/dev/null 2>&1 &
echo "Client started in background (PID $!)."
