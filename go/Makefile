.PHONY: build run test clean deps fmt lint docker-build docker-run

# Build the blockchain node
build:
	go build -o bin/oxy-blockchain ./cmd/oxy-blockchain/main.go

# Run the blockchain node
run: build
	./bin/oxy-blockchain

# Run tests (ejecuta desde test/scripts/)
test:
	@echo "Ejecutando tests desde test/scripts/test.bat..."
	@if exist test\scripts\test.bat (call test\scripts\test.bat) else (go test ./... -v)

# Run tests with coverage
test-coverage:
	go test ./... -coverprofile=test/coverage.out
	go tool cover -html=test/coverage.out -o test/coverage.html
	@echo "Coverage reporte generado: test/coverage.html"

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf data/
	rm -rf coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requiere golangci-lint)
lint:
	golangci-lint run ./...

# Docker build
docker-build:
	docker build -t oxy-blockchain:latest -f ../Dockerfile ../go

# Docker run (development)
docker-run:
	docker run -it --rm \
		-p 8080:8080 \
		-e OXY_DATA_DIR=/app/data \
		-e OXY_LOG_LEVEL=info \
		-e BLOCKCHAIN_API_ENABLED=true \
		-e BLOCKCHAIN_API_PORT=8080 \
		-e BLOCKCHAIN_API_HOST=0.0.0.0 \
		-v $(PWD)/data:/app/data \
		oxy-blockchain:latest

# Run integration tests
test-integration:
	go test ./internal/consensus -run TestABCIApp -v
	go test ./internal/crypto -run TestVerify -v
