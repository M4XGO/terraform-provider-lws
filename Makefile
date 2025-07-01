default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Generate or update documentation
.PHONY: docs
docs:
	@echo "Generating documentation with tfplugindocs..."
	go generate ./...
	@echo "Formatting example terraform files..."
	terraform fmt -recursive ./examples/ 2>/dev/null || echo "No examples directory found"

# Build the provider
.PHONY: build
build:
	go build -o terraform-provider-lws

# Install the provider locally
.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/maximenony/lws/0.1.0/darwin_amd64
	cp terraform-provider-lws ~/.terraform.d/plugins/registry.terraform.io/maximenony/lws/0.1.0/darwin_amd64/

# Clean build artifacts
.PHONY: clean
clean:
	rm -f terraform-provider-lws

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# Test the provider (unit tests only)
.PHONY: test
test:
	go test ./internal/provider -v

# Test with coverage
.PHONY: test-coverage
test-coverage:
	go test ./internal/provider -v -coverprofile=coverage.out
	go tool cover -func=coverage.out

# Generate HTML coverage report
.PHONY: test-coverage-html
test-coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Test specific patterns
.PHONY: test-unit
test-unit:
	go test ./internal/provider -v -run="Test.*Unit|Test.*Client|Test.*DataSource|Test.*Resource.*Metadata|Test.*Resource.*Schema"

# Test integration
.PHONY: test-integration
test-integration:
	go test ./internal/provider -v -run="Test.*Provider.*Complete|Test.*Provider.*Error|Test.*Provider.*Auth|Test.*Integration"

# Test validation logic
.PHONY: test-validation
test-validation:
	go test ./internal/provider -v -run="Test.*Validation|Test.*Types|Test.*TTL|Test.*Value"

# All tests including acceptance (requires credentials)
.PHONY: test-all
test-all:
	go test ./... -v

# Download dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Verify dependencies
.PHONY: verify
verify:
	go mod verify

# Security scan is now integrated in the main lint target with gosec enabled

# Lint and fix auto-fixable issues
.PHONY: lint-fix
lint-fix:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix

# Install development tools
.PHONY: tools
tools:
	go install github.com/goreleaser/goreleaser@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Validate GoReleaser configuration
.PHONY: release-check
release-check:
	goreleaser check

# Test release build (snapshot)
.PHONY: release-test
release-test:
	goreleaser build --snapshot --clean

# Full CI workflow locally
.PHONY: ci
ci: deps verify fmt lint test-coverage

# Development workflow
.PHONY: dev
dev: deps fmt lint test

# Pre-commit hooks
.PHONY: pre-commit
pre-commit: fmt lint test-unit

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build             - Build the provider binary"
	@echo "  test              - Run unit tests"
	@echo "  test-coverage     - Run tests with coverage"
	@echo "  test-integration  - Run integration tests"
	@echo "  test-validation   - Run validation tests"
	@echo "  test-all          - Run all tests including acceptance"
	@echo "  testacc           - Run acceptance tests (requires credentials)"
	@echo "  lint              - Run linters"
	@echo "  lint-fix          - Run linters and auto-fix issues"
	@echo "  security          - Run security scans"
	@echo "  fmt               - Format code"
	@echo "  install           - Install provider locally"
	@echo "  release-check     - Validate GoReleaser config"
	@echo "  release-test      - Test release build"
	@echo "  ci                - Run full CI workflow"
	@echo "  dev               - Development workflow"
	@echo "  clean             - Clean build artifacts"
	@echo "  tools             - Install development tools"
	@echo "  deps              - Download dependencies"
	@echo "  verify            - Verify dependencies"
	@echo "  help              - Show this help" 