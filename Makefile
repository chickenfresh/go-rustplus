.PHONY: all build test clean proto lint build-windows build-linux

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Binary names
RUSTPLUS_CLI=rustplus-cli
PUSH_RECEIVER_CLI=go-rustplus
CMD_RUSTPLUS_DIR=./cmd/rustplus
CMD_PUSH_RECEIVER_DIR=./cmd/push-receiver

# Output directories
BIN_DIR=./bin
BIN_LINUX=$(BIN_DIR)/linux
BIN_WINDOWS=$(BIN_DIR)/windows

# Protocol buffer parameters
PROTOC=protoc
FCM_PROTO_DIR=./fcm/proto
RUSTPLUS_PROTO_DIR=./rustplus/proto
FCM_PROTO_FILES=mcs.proto checkin.proto android_checkin.proto
RUSTPLUS_PROTO_FILES=$(RUSTPLUS_PROTO_DIR)/rustplus.proto

# Examples directory
EXAMPLES_DIR=./examples
EXAMPLES_BIN_DIR=$(BIN_DIR)/examples

# Main build target
all: proto build build-examples

# Create output directories
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(BIN_LINUX):
	mkdir -p $(BIN_LINUX)

$(BIN_WINDOWS):
	mkdir -p $(BIN_WINDOWS)

# Create examples output directory
$(EXAMPLES_BIN_DIR):
	mkdir -p $(EXAMPLES_BIN_DIR)

# Build all applications
build: $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(RUSTPLUS_CLI) $(CMD_RUSTPLUS_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(PUSH_RECEIVER_CLI) $(CMD_PUSH_RECEIVER_DIR)

# Build for Linux
build-linux: $(BIN_LINUX)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BIN_LINUX)/$(RUSTPLUS_CLI) $(CMD_RUSTPLUS_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BIN_LINUX)/$(PUSH_RECEIVER_CLI) $(CMD_PUSH_RECEIVER_DIR)

# Build for Windows
build-windows: $(BIN_WINDOWS)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BIN_WINDOWS)/$(RUSTPLUS_CLI).exe $(CMD_RUSTPLUS_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BIN_WINDOWS)/$(PUSH_RECEIVER_CLI).exe $(CMD_PUSH_RECEIVER_DIR)

# Build only the Rust+ CLI application
build-rustplus: $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(RUSTPLUS_CLI) $(CMD_RUSTPLUS_DIR)

# Build only the Push Receiver CLI application
build-push-receiver: $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(PUSH_RECEIVER_CLI) $(CMD_PUSH_RECEIVER_DIR)

# Build examples
build-examples: $(EXAMPLES_BIN_DIR)
	@echo "Building examples..."
	@for dir in $(shell find $(EXAMPLES_DIR) -type d -mindepth 1 -maxdepth 1); do \
		example_name=$$(basename $$dir); \
		echo "Building $$example_name..."; \
		$(GOBUILD) -o $(EXAMPLES_BIN_DIR)/$$example_name $$dir; \
	done

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	rm -f $(RUSTPLUS_CLI) $(PUSH_RECEIVER_CLI)

# Generate all protocol buffer code
proto: proto-fcm proto-rustplus

# Generate FCM protocol buffer code
proto-fcm:
	cd $(FCM_PROTO_DIR) && $(PROTOC) --go_out=. --go_opt=paths=source_relative $(FCM_PROTO_FILES)

# Generate Rust+ protocol buffer code
proto-rustplus:
	$(PROTOC) --go_out=. --go_opt=paths=source_relative $(RUSTPLUS_PROTO_FILES)

# Install dependencies
deps:
	$(GOMOD) download
	$(GOGET) google.golang.org/protobuf/cmd/protoc-gen-go

# Run linter
lint:
	$(GOLINT) run

# Install the binaries to GOPATH/bin
install:
	$(GOBUILD) -o $(GOPATH)/bin/$(RUSTPLUS_CLI) $(CMD_RUSTPLUS_DIR)
	$(GOBUILD) -o $(GOPATH)/bin/$(PUSH_RECEIVER_CLI) $(CMD_PUSH_RECEIVER_DIR)

# Run the Rust+ application
run-rustplus: build-rustplus
	$(BIN_DIR)/$(RUSTPLUS_CLI)

# Run the Push Receiver application
run-push-receiver: build-push-receiver
	$(BIN_DIR)/$(PUSH_RECEIVER_CLI)

# Generate protocol buffers and build for release
release: proto build build-linux build-windows

# Help target
help:
	@echo "Available targets:"
	@echo "  all               - Generate protocol buffers and build all applications"
	@echo "  build             - Build all applications to bin directory"
	@echo "  build-linux       - Build applications for Linux"
	@echo "  build-windows     - Build applications for Windows"
	@echo "  build-rustplus    - Build only the Rust+ CLI application"
	@echo "  build-push-receiver - Build only the Push Receiver CLI application"
	@echo "  test              - Run tests"
	@echo "  clean             - Clean build artifacts"
	@echo "  proto             - Generate all protocol buffer code"
	@echo "  proto-fcm         - Generate only FCM protocol buffer code"
	@echo "  proto-rustplus    - Generate only Rust+ protocol buffer code"
	@echo "  deps              - Install dependencies"
	@echo "  lint              - Run linter"
	@echo "  install           - Install the binaries to GOPATH/bin"
	@echo "  run-rustplus      - Build and run the Rust+ application"
	@echo "  run-push-receiver - Build and run the Push Receiver application"
	@echo "  release           - Generate protocol buffers and build for all platforms"
	@echo "  build-examples    - Build all example applications" 