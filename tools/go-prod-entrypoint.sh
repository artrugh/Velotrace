#!/bin/sh

set -e

# Wait for database to be ready
until pg_isready -h db -p 5432; do
    echo "Waiting for database..."
    sleep 2
done

echo "Running database migrations..."
goose -dir internal/db/sql postgres "$DATABASE_URL" up

echo "🚀 Starting Production API..."
exec ./main
