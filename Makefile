.PHONY: run build build-web test test-web dev dev-web docker-up docker-down clean

# Build API-only binary
build:
	cd api && go build -o ../paymob-api ./cmd/server/

# Build with web frontend (HTMX)
build-web:
	cd api && go build -tags web -o ../paymob-full ./cmd/server/

# Run API-only
run: build
	./paymob-api

# Run with web frontend
run-web: build-web
	./paymob-full

# Development mode - API only
dev:
	cd api && go run ./cmd/server/

# Development mode - with web frontend
dev-web:
	cd api && go run -tags web ./cmd/server/

# Run web frontend standalone (separate terminal)
dev-web-frontend:
	cd web && go run ./cmd/server/

# Run tests
test:
	cd api && go test -tags web ./...

# Run tests with coverage
test-coverage:
	cd api && go test -tags web -coverprofile=../coverage.out ./...
	go tool cover -func=coverage.out | tail -1

# Run benchmarks
bench:
	cd api && go test -tags web -bench=. -benchmem ./...

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
	rm -f paymob-api paymob-full

# Show help
help:
	@echo "PayMob Demo - Available commands:"
	@echo "  make build          - Build API-only binary"
	@echo "  make build-web      - Build with HTMX web frontend"
	@echo "  make run            - Build and run API-only"
	@echo "  make run-web        - Build and run with web frontend"
	@echo "  make dev            - Run API-only (development)"
	@echo "  make dev-web        - Run with web frontend (development)"
	@echo "  make dev-web-frontend - Run standalone web frontend"
	@echo "  make test           - Run tests"
	@echo "  make bench          - Run benchmarks"
	@echo "  make docker-up      - Start with Docker"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make clean          - Remove build artifacts"
