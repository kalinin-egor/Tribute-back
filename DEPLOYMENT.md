# Deployment Guide

## Overview
This guide explains how to deploy the Tribute Backend application with proper database migrations.

## Prerequisites
- Docker and Docker Compose installed
- Git repository access
- Server with SSH access (for production)

## Local Development

### 1. Clone and Setup
```bash
git clone <repository-url>
cd Tribute-back
```

### 2. Environment Setup
```bash
cp env.example .env
# Edit .env with your local settings
```

### 3. Start Services with Migrations
```bash
# Option 1: Full setup with migrations
make dev-setup

# Option 2: Manual step-by-step
docker-compose up -d postgres redis
sleep 10  # Wait for PostgreSQL to be ready
make migrate-up-docker
docker-compose up -d app
```

### 4. Verify Deployment
```bash
# Check service status
docker-compose ps

# Check migration status
make migrate-status-docker

# Test API endpoints
curl http://localhost:8081/health
curl http://localhost:8081/docs/index.html
```

## Production Deployment

### CI/CD Pipeline
The GitHub Actions workflow automatically:
1. Builds the application
2. Deploys to your VPS
3. Starts Docker containers
4. **Runs database migrations** (newly added)
5. Verifies deployment

### Manual Production Deployment

If you need to deploy manually:

```bash
# 1. SSH to your server
ssh root@your-server-ip

# 2. Navigate to project directory
cd /root/tribute/Tribute-back

# 3. Pull latest changes
git pull origin main

# 4. Stop existing services
docker-compose down

# 5. Start services
docker-compose up -d

# 6. Run migrations (if not done automatically)
./run-migrations.sh
# or
make migrate-up-docker
```

## Database Migrations

### Understanding Migrations
- **001_create_users_table.up.sql**: Creates users table with Telegram User ID
- **002_create_tribute_tables.up.sql**: Creates channels, subscriptions, payments tables
- **003_add_onboarding_to_users.up.sql**: Deprecated (field already in 001)

### Migration Commands

```bash
# Check migration status
make migrate-status-docker

# Run all pending migrations
make migrate-up-docker

# Rollback last migration
make migrate-down-docker

# Force specific migration version
make migrate-force-docker

# Manual migration script
./run-migrations.sh
```

### Troubleshooting Migrations

#### Error: "column user_id does not exist"
This error occurs when migrations haven't been applied. Solutions:

1. **Check if migrations ran:**
   ```bash
   make migrate-status-docker
   ```

2. **Force run migrations:**
   ```bash
   make migrate-up-docker
   ```

3. **Reset database (WARNING: loses data):**
   ```bash
   docker-compose down -v
   docker-compose up -d postgres
   sleep 10
   make migrate-up-docker
   ```

#### Error: "relation users does not exist"
The database is completely empty. Run:
```bash
./run-migrations.sh
```

## Environment Variables

### Required for Production
```bash
# Database
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=your-db-user
DB_PASSWORD=your-db-password
DB_NAME=your-db-name
DB_SSL_MODE=require

# Server
PORT=8081
ENV=production

# CORS
ALLOWED_ORIGINS=https://your-frontend-domain.com
```

## Monitoring

### Health Checks
```bash
# Application health
curl http://your-domain:8081/health

# Database connection
docker-compose exec postgres pg_isready -U postgres

# Redis connection
docker-compose exec redis redis-cli ping
```

### Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f app
docker-compose logs -f postgres
```

## Rollback

### Application Rollback
```bash
# Revert to previous commit
git reset --hard HEAD~1
docker-compose up -d --build
```

### Database Rollback
```bash
# Rollback last migration
make migrate-down-docker

# Rollback to specific version
make migrate-force-docker
```

## Security Notes

1. **Never commit `.env` files** with production credentials
2. **Use strong passwords** for database and Redis
3. **Enable SSL** for database connections in production
4. **Restrict CORS origins** to your frontend domains only
5. **Use secrets management** for sensitive data in CI/CD

## Support

If you encounter issues:
1. Check logs: `docker-compose logs -f`
2. Verify migrations: `make migrate-status-docker`
3. Test database connection: `docker-compose exec postgres psql -U postgres -d tribute_db`
4. Check service health: `docker-compose ps` 