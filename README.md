# Tribute Backend API

A clean, well-structured Golang REST API backend built with Gin framework, PostgreSQL database, Redis caching, and JWT authentication.

## Architecture

This project follows a clean architecture pattern with clear separation of concerns:

```
├── cmd/                    # Application entry points
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── database/         # Database connection and setup
│   ├── redis/            # Redis connection and caching
│   ├── handler/          # HTTP request handlers
│   ├── middleware/       # HTTP middleware (auth, CORS, etc.)
│   ├── models/           # Data models and DTOs
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic layer
│   └── server/           # HTTP server setup
├── migrations/           # Database migrations
├── docker-compose.yml    # Docker services configuration
├── main.go              # Application entry point
├── go.mod               # Go module file
├── Makefile             # Build and development commands
└── README.md            # This file
```

## Features

- **Clean Architecture**: Separation of concerns with layers (Handler → Service → Repository → Database)
- **JWT Authentication**: Secure token-based authentication
- **PostgreSQL Database**: Robust relational database with migrations
- **Redis Caching**: Fast in-memory caching and session storage
- **Docker Compose**: Easy development environment setup
- **CORS Support**: Cross-origin resource sharing configuration
- **Environment Configuration**: Flexible configuration management
- **Password Hashing**: Secure password storage with bcrypt
- **Input Validation**: Request validation with Gin binding
- **Error Handling**: Consistent error responses
- **Database Migrations**: Version-controlled database schema changes

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional, for using Makefile commands)

## Quick Start with Docker Compose

### 1. **Clone the repository**
```bash
git clone <repository-url>
cd Tribute-back
```

### 2. **Set up environment variables**
```bash
cp env.example .env
# Edit .env if needed (defaults work with Docker Compose)
```

### 3. **Start services with Docker Compose**
```bash
make dev-setup
# or manually:
docker-compose up -d
```

### 4. **Run the application**
```bash
make run
# or manually:
go run main.go
```

## Manual Setup (without Docker)

### 1. **Install dependencies**
```bash
make deps
# or manually:
go mod download
go mod tidy
```

### 2. **Set up PostgreSQL and Redis**
```bash
# Start PostgreSQL (if using Homebrew)
brew services start postgresql@15

# Start Redis
brew services start redis

# Create database
createdb tribute_db
```

### 3. **Set up environment variables**
```bash
cp env.example .env
# Edit .env with your database credentials
```

### 4. **Run database migrations**
```bash
make install-migrate
make migrate-up
```

### 5. **Run the application**
```bash
make run
```

## Environment Variables

Create a `.env` file based on `env.example`:

```env
# Server Configuration
PORT=8080
ENV=development

# Database Configuration (Docker Compose)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=tribute_db
DB_SSL_MODE=disable

# Redis Configuration (Docker Compose)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY=24h

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user

### User Management (Protected)
- `GET /api/v1/users/profile` - Get current user profile
- `PUT /api/v1/users/profile` - Update current user profile
- `GET /api/v1/users/:id` - Get user by ID
- `GET /api/v1/users/` - List all users (with pagination)
- `DELETE /api/v1/users/:id` - Delete user

### Health Check
- `GET /health` - Health check endpoint

## API Examples

### Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "username",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Get Profile (with JWT token)
```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Development

### Available Make Commands

#### Docker Commands
- `make docker-up` - Start Docker services
- `make docker-down` - Stop Docker services
- `make docker-logs` - View Docker logs
- `make docker-restart` - Restart Docker services

#### Application Commands
- `make build` - Build the application
- `make run` - Run the application
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make deps` - Install dependencies

#### Database Commands
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback database migrations
- `make migrate-create` - Create new migration
- `make install-migrate` - Install migrate tool

#### Development Setup
- `make dev-setup` - Complete development setup with Docker
- `make dev-setup-local` - Development setup without Docker
- `make dev-full` - Full development workflow

### Hot Reload (Optional)

Install Air for hot reload during development:

```bash
make install-air
make dev
```

## Docker Services

The `docker-compose.yml` includes:

- **PostgreSQL 15**: Main database
- **Redis 7**: Caching and session storage
- **Migrate**: Database migration tool

### Service URLs
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- API: `localhost:8080`

## Database Migrations

The project uses `golang-migrate` for database migrations:

```bash
# With Docker Compose
docker-compose exec migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" up

# Local
make migrate-up
```

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...
```

## Production Deployment

1. Set environment variables for production
2. Build the application: `make build`
3. Run database migrations
4. Start the application with proper process management

## Contributing

1. Follow the existing code structure and patterns
2. Add tests for new features
3. Update documentation as needed
4. Use conventional commit messages

## License

This project is licensed under the MIT License. 