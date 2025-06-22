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

# Run migrations
echo "📋 Executing migrations..."
docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" up

echo "✅ Migrations completed successfully!"

# Show migration status
echo "📊 Migration status:"
docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" version 