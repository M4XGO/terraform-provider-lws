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

# Test the provider
.PHONY: test
test:
	go test ./...

# Download dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy 