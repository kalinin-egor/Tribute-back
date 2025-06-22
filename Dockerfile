# Build stage
FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Install swag utility
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

# Generate swagger docs. The binary is located at /go/bin/swag.
RUN /go/bin/swag init

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM debian:stable-slim

WORKDIR /root/

# Install ca-certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/main .
# Copy the freshly generated docs
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/.env ./.env

EXPOSE 8080

CMD ["./main"]
