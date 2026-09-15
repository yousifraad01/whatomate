.PHONY: all build build-prod embed-frontend run run-migrate migrate \
	test test-race test-coverage vet fmt fmt-check lint check \
	clean deps deps-update \
	docker-build docker-up docker-down docker-logs docker-restart dev-deps dev-deps-down \
	frontend-install frontend-dev frontend-build frontend-preview frontend-check \
	dev swagger help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=whatomate
BINARY_PATH=./cmd/whatomate
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# All Go packages share one test database, so packages must not run in
# parallel (mirrors .github/workflows/test.yml). TEST_TIMEOUT caps a hung run.
TEST_PKGS=$(shell $(GOCMD) list ./... | grep -v /test/)
TEST_FLAGS?=-p 1 -timeout 15m
GO_SRC_DIRS=./cmd ./internal ./pkg ./test

# Docker parameters
DOCKER_COMPOSE=docker compose -f docker/docker-compose.yml

all: build

# Build the backend (development - without frontend)
build:
	$(GOBUILD) -o $(BINARY_NAME) $(BINARY_PATH)

# Build production binary with embedded frontend.
# embed-frontend depends on frontend-build, so `make -j build-prod` cannot
# copy the bundle before it has been produced.
build-prod: embed-frontend
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(BINARY_PATH)
	@echo "Production binary built: $(BINARY_NAME)"
	@echo "Version: $(VERSION)"
	@ls -lh $(BINARY_NAME)

# Copy frontend build to embed directory (keeps the .gitkeep placeholder)
embed-frontend: frontend-build
	@echo "Copying frontend build to embed directory..."
	@rm -rf internal/frontend/dist/*
	@cp -r frontend/dist/. internal/frontend/dist/
	@echo "Frontend embedded successfully"

# Run the backend locally
run:
	$(GOCMD) run $(BINARY_PATH)/main.go server -config config.toml

# Run with migrations
run-migrate:
	$(GOCMD) run $(BINARY_PATH)/main.go server -config config.toml -migrate

# Database migrations
migrate: run-migrate

# Run tests. Uses gotestsum when available for live progress + a clear
# failure summary at the end. Falls back to the built-in `go test -v` so
# nothing breaks for devs who haven't installed it.
# Install:  go install gotest.tools/gotestsum@latest
# Set TEST_DATABASE_URL / TEST_REDIS_URL to run the database-backed tests
# (see `make dev-deps`); without them those tests are skipped.
test:
	@if command -v gotestsum >/dev/null 2>&1; then \
		gotestsum --format testname --hide-summary=skipped -- $(TEST_FLAGS) $(TEST_PKGS); \
	else \
		echo "(install gotestsum for nicer output: go install gotest.tools/gotestsum@latest)"; \
		$(GOTEST) -v $(TEST_FLAGS) $(TEST_PKGS); \
	fi

# Run tests with the race detector (requires cgo; on Windows this needs a C
# toolchain such as mingw-w64, so it is easier to run in the CI/Linux image).
test-race:
	CGO_ENABLED=1 $(GOTEST) -race $(TEST_FLAGS) $(TEST_PKGS)

# Run tests with coverage. Same gotestsum fallback as `make test`.
test-coverage:
	@if command -v gotestsum >/dev/null 2>&1; then \
		gotestsum --format testname --hide-summary=skipped -- $(TEST_FLAGS) -coverprofile=coverage.out $(TEST_PKGS); \
	else \
		echo "(install gotestsum for nicer output: go install gotest.tools/gotestsum@latest)"; \
		$(GOTEST) -v $(TEST_FLAGS) -coverprofile=coverage.out $(TEST_PKGS); \
	fi
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Static analysis
vet:
	$(GOCMD) vet ./...

# Format code
fmt:
	$(GOCMD) fmt ./...

# Fail if any file is not gofmt-clean (used by `make check`)
fmt-check:
	@unformatted="$$(gofmt -l $(GO_SRC_DIRS))"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:"; echo "$$unformatted"; exit 1; \
	fi

# Lint (matches the version pinned in .github/workflows/test.yml)
lint:
	golangci-lint run ./...

# Everything CI runs for the backend, in one target
check: fmt-check vet test

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
deps-update:
	$(GOMOD) tidy
	$(GOGET) -u ./...

# Docker commands
docker-build:
	$(DOCKER_COMPOSE) build

docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f

docker-restart:
	$(DOCKER_COMPOSE) restart

# Start only PostgreSQL and Redis for local development / tests
dev-deps:
	$(DOCKER_COMPOSE) up -d db redis
	@echo "Postgres on 127.0.0.1:5432, Redis on 127.0.0.1:6379"
	@echo "For tests: export TEST_DATABASE_URL=postgres://whatomate:whatomate@127.0.0.1:5432/whatomate?sslmode=disable TEST_REDIS_URL=redis://127.0.0.1:6379/1  (use a Redis DB the dev server does not, or its worker consumes queue-test jobs)"

dev-deps-down:
	$(DOCKER_COMPOSE) stop db redis

# Frontend commands
frontend-install:
	cd frontend && npm ci

frontend-dev:
	@if [ ! -d "frontend/node_modules" ]; then \
		echo "Installing frontend dependencies..."; \
		cd frontend && npm ci; \
	fi
	cd frontend && npm run dev

frontend-build:
	@if [ ! -d "frontend/node_modules" ]; then \
		echo "Installing frontend dependencies..."; \
		cd frontend && npm ci; \
	fi
	cd frontend && npm run build

frontend-preview:
	cd frontend && npm run preview

# Type check, lint (without auto-fix) and unit tests
frontend-check:
	cd frontend && npm run typecheck
	cd frontend && npx eslint . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts --ignore-path .gitignore
	cd frontend && npm run test:unit

# Development - run backend and frontend together; Ctrl-C stops both and a
# failure of either ends the session instead of leaving an orphan running.
dev:
	@echo "Starting backend and frontend in development mode..."
	@trap 'kill 0' EXIT INT TERM; \
	$(MAKE) run-migrate & \
	$(MAKE) frontend-dev & \
	wait

# Generate swagger docs (if using)
swagger:
	swag init -g cmd/whatomate/main.go -o api/docs

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Production:"
	@echo "  build-prod     - Build single binary with embedded frontend"
	@echo ""
	@echo "Development:"
	@echo "  build          - Build the backend binary (without frontend)"
	@echo "  run            - Run the backend locally"
	@echo "  run-migrate    - Run the backend with database migrations"
	@echo "  dev            - Run both backend and frontend in development mode"
	@echo "  dev-deps       - Start only Postgres and Redis in Docker"
	@echo "  dev-deps-down  - Stop Postgres and Redis"
	@echo ""
	@echo "Frontend:"
	@echo "  frontend-install - Install frontend dependencies (npm ci)"
	@echo "  frontend-dev   - Run frontend in development mode"
	@echo "  frontend-build - Build frontend for production"
	@echo "  frontend-check - Type check, lint and unit-test the frontend"
	@echo ""
	@echo "Testing & quality:"
	@echo "  test           - Run tests (packages run sequentially, shared test DB)"
	@echo "  test-race      - Run tests with the race detector (needs cgo)"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  vet            - Run go vet"
	@echo "  fmt / fmt-check - Format code / verify formatting"
	@echo "  lint           - Run golangci-lint"
	@echo "  check          - fmt-check + vet + test"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-up      - Start Docker containers"
	@echo "  docker-down    - Stop Docker containers"
	@echo "  docker-logs    - View Docker logs"
	@echo ""
	@echo "Other:"
	@echo "  clean          - Remove build artifacts"
	@echo "  deps           - Download dependencies"
