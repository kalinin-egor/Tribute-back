#!/bin/bash

# Script to run database migrations manually
# Usage: ./run-migrations.sh

set -e

echo "🗄️ Running database migrations..."

# Check if docker-compose is running
if ! docker-compose ps | grep -q "postgres.*Up"; then
    echo "❌ PostgreSQL container is not running. Starting services..."
    docker-compose up -d postgres
    echo "⏳ Waiting for PostgreSQL to be ready..."
    sleep 10
fi

# Check migration status and fix dirty state if needed
echo "📊 Checking migration status..."
MIGRATION_STATUS=$(docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" version 2>&1 || echo "ERROR")

if [[ "$MIGRATION_STATUS" == *"dirty"* ]]; then
    echo "⚠️ Database is in dirty state. Attempting to fix..."
    DIRTY_VERSION=$(echo "$MIGRATION_STATUS" | grep -o '[0-9]\+' | head -1)
    echo "🔧 Forcing migration version to $DIRTY_VERSION..."
    docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" force $DIRTY_VERSION
    echo "✅ Dirty state fixed!"
fi

# Run migrations
echo "📋 Executing migrations..."
docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" up

echo "✅ Migrations completed successfully!"

# Show migration status
echo "📊 Migration status:"
docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" version 