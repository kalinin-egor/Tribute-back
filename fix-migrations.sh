#!/bin/bash

# Script to automatically fix migration issues
# Usage: ./fix-migrations.sh

set -e

echo "🔧 Migration Auto-Fix Script"
echo "============================"

# Check if docker-compose is running
if ! docker-compose ps | grep -q "postgres.*Up"; then
    echo "❌ PostgreSQL container is not running. Starting services..."
    docker-compose up -d postgres
    echo "⏳ Waiting for PostgreSQL to be ready..."
    sleep 10
fi

# Function to check migration status
check_migration_status() {
    echo "📊 Checking migration status..."
    MIGRATION_STATUS=$(docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" version 2>&1 || echo "ERROR")
    echo "Status: $MIGRATION_STATUS"
    echo "$MIGRATION_STATUS"
}

# Function to fix dirty state
fix_dirty_state() {
    local status="$1"
    if [[ "$status" == *"dirty"* ]]; then
        echo "⚠️ Database is in dirty state. Attempting to fix..."
        # Extract version number from dirty status
        DIRTY_VERSION=$(echo "$status" | grep -o '[0-9]\+' | head -1)
        echo "🔧 Forcing migration version to $DIRTY_VERSION..."
        docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" force $DIRTY_VERSION
        echo "✅ Dirty state fixed!"
        return 0
    fi
    return 1
}

# Function to run migrations
run_migrations() {
    echo "📋 Running migrations..."
    docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" up
    echo "✅ Migrations completed successfully!"
}

# Main execution
echo "🚀 Starting migration fix process..."

# Step 1: Check current status
STATUS=$(check_migration_status)

# Step 2: Fix dirty state if needed
if fix_dirty_state "$STATUS"; then
    echo "🔄 Re-checking status after fix..."
    STATUS=$(check_migration_status)
fi

# Step 3: Run migrations
run_migrations

# Step 4: Final status check
echo "📊 Final migration status:"
docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" version

echo "🎉 Migration fix process completed!" 