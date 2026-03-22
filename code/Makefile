.PHONY: all build test test-unit test-integration lint clean run docker-build docker-run

APP_NAME := swarm-predict
BUILD_DIR := ./build
DOCKER_IMAGE := mirofish/swarm-predict
DOCKER_TAG := latest

all: lint test build

## Build
build:
	@echo "==> Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

## Run
run: build
	@echo "==> Running $(APP_NAME)..."
	$(BUILD_DIR)/$(APP_NAME)

## Test
test: test-unit test-integration

test-unit:
	@echo "==> Running unit tests..."
	go test -v -race -count=1 ./internal/...

test-integration:
	@echo "==> Running integration tests..."
	go test -v -race -count=1 -tags=integration ./test/integration/...

test-cover:
	@echo "==> Running tests with coverage..."
	go test -race -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Lint
lint:
	@echo "==> Running go vet..."
	go vet ./...

## Clean
clean:
	@echo "==> Cleaning..."
	rm -rf $(BUILD_DIR) coverage.out coverage.html

## Docker
docker-build:
	@echo "==> Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build
	@echo "==> Running Docker container..."
	docker run --rm -p 8080:8080 \
		-e OPENCLAW_GATEWAY_URL=http://host.docker.internal:18789 \
		-e OPENCLAW_API_KEY=$${OPENCLAW_API_KEY} \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

## Dependencies
deps:
	@echo "==> Downloading dependencies..."
	go mod download
	go mod tidy

## Help
help:
	@echo "Available targets:"
	@echo "  all              - lint, test, build"
	@echo "  build            - compile the binary"
	@echo "  run              - build and run locally"
	@echo "  test             - run all tests"
	@echo "  test-unit        - run unit tests only"
	@echo "  test-integration - run integration tests only"
	@echo "  test-cover       - run tests with coverage report"
	@echo "  lint             - run go vet"
	@echo "  clean            - remove build artifacts"
	@echo "  docker-build     - build Docker image"
	@echo "  docker-run       - build and run Docker container"
	@echo "  deps             - download and tidy dependencies"
