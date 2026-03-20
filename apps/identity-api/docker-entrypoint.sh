#!/bin/sh

set -e

echo "Running database migrations..."
goose -dir /migrations postgres "$DATABASE_URL" up

echo "Starting Identity API..."
exec /velotrace-identity
