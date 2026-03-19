.PHONY: run build test docker-up docker-down clean

# Build the application
build:
	go build -o paymob-demo ./cmd/server

# Run the application
run: build
	./paymob-demo

# Run in development mode
dev:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Docker build
docker-build:
	docker build -t paymob-demo .

# Docker run
docker-up:
	docker-compose up -d

# Docker stop
docker-down:
	docker-compose down

# Clean build artifacts
clean:
	rm -f paymob-demo

# Show help
help:
	@echo "PayMob Demo - Available commands:"
	@echo "  make build       - Build the Go application"
	@echo "  make run         - Build and run the application"
	@echo "  make dev         - Run in development mode"
	@echo "  make test        - Run unit tests"
	@echo "  make bench       - Run performance benchmarks"
	@echo "  make docker-up   - Start with Docker"
	@echo "  make docker-down - Stop Docker containers"
	@echo "  make clean       - Remove build artifacts"
