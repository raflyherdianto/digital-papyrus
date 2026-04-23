#!/bin/sh
set -e

echo "Ensuring required directories exist..."
mkdir -p /app/data
mkdir -p /app/frontend/public/uploads

echo "Fixing permissions for mounted volumes..."
chown -R appuser:appgroup /app/data
chown -R appuser:appgroup /app/frontend/public/uploads

echo "Starting the application as appuser..."
exec su-exec appuser "$@"
