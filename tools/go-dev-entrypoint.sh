#!/bin/sh

set -e

# Wait for database to be ready
until pg_isready -h db -p 5432; do
    echo "Waiting for database..."
    sleep 2
done

echo "Running database migrations..."
goose -dir internal/db/sql postgres "$DATABASE_URL" up

if [ "$GO_ENV" = "development" ]; then
    echo "🌱 Seeding database..."
    go run ./internal/db/seed/main.go
fi

echo "✨ Starting API with hot-reload..."
exec air -c .air.toml
