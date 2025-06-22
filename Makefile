.PHONY: build run test clean deps swagger docker-up docker-down docker-logs docker-restart docker-build docker-run

# Build the application
build:
	go build -o bin/tribute-back main.go

# Run the application
run:
	go run main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Install swag tool for Swagger generation
install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
swagger:
	swag init -g main.go

# Generate Swagger and run
swagger-run: swagger run

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart:
	docker-compose restart

# Build and run with Docker
docker-build:
	docker-compose build

docker-run: docker-build docker-up
	@echo "Application is running at http://localhost:8080"
	@echo "Swagger docs: http://localhost:8080/docs/index.html"
	@echo "Health check: http://localhost:8080/health"

# Development with Docker (full stack)
dev-docker: docker-run
	@echo "Full development environment ready!"
	@echo "API: http://localhost:8080"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"

# Development setup with Docker
dev-setup: deps docker-up
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Running migrations..."
	@docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" up
	@echo "Development environment setup complete!"

# Development setup without Docker
dev-setup-local: deps
	@echo "Development environment setup complete!"

# Run with hot reload (requires air)
dev:
	air

# Install air for hot reload
install-air:
	go install github.com/cosmtrek/air@latest

# Full development workflow
dev-full: docker-up
	@echo "Starting development environment..."
	@sleep 5
	@echo "Starting application..."
	@go run main.go

# Run database migrations
migrate-up:
	docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" up

# Check migration status
migrate-status:
	docker-compose exec -T migrate migrate -path /migrations -database "postgres://postgres:password@postgres:5432/tribute_db?sslmode=disable" version