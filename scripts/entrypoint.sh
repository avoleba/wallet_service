#!/bin/sh
set -e
if [ -n "$DATABASE_DSN" ] && command -v migrate >/dev/null 2>&1; then
  echo "Running migrations..."
  migrate -path /migrations -database "$DATABASE_DSN" up
fi
exec ./wallet-api
