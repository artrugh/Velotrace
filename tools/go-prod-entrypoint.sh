#!/bin/sh

set -e

# Run migrations using Goose
if [ -d "/migrations" ]; then
    echo "🚀 Running database migrations..."
    goose -dir /migrations postgres "$DATABASE_URL" up
else
    echo "ℹ️ No migrations directory found, skipping..."
fi

# Start the application
# We expect the binary path to be passed as the first argument or use a default
echo "✨ Starting Application..."
exec "$@"
