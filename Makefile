default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Generate or update documentation
.PHONY: docs
docs:
	go generate ./...

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
	golangci-lint run

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