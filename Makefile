# Project variables
IMAGE_NAME := vddk
REGISTRY := quay.io/yaacov
TAG := latest
DEPLOY_YAML := deployment.yaml
BUILD_DIR := .

# Initialize the Go project for the first time
.PHONY: init
goinit:
	@echo "Initializing Go project..."
	go mod init vddk-builder || echo "go.mod already exists"
	go mod tidy
	@echo "Go project initialized successfully!"

# Generate self-signed certificates for local testing
.PHONY: gen-certs
gen-certs:
	@echo "Generating self-signed certificates for local testing..."
	mkdir -p certs
	openssl req -x509 -nodes -newkey rsa:4096 -keyout certs/server.key \
		-out certs/server.crt -days 365 -subj "/CN=localhost"
	@echo "Certificates generated at ./certs/server.crt and ./certs/server.key."

# Run a local registry with Podman (no authentication)
.PHONY: run-registry
run-registry:
	@echo "Starting a local Podman registry on port 5000 without authentication..."
	podman run -d --name local-registry -p 5000:5000 --restart=always \
		-v $(PWD)/certs:/certs:Z \
		-e "REGISTRY_HTTP_TLS_CERTIFICATE=/certs/server.crt" \
		-e "REGISTRY_HTTP_TLS_KEY=/certs/server.key" registry:2
	@echo "Local registry is running at https://localhost:5000."

# Stop and remove the local registry
.PHONY: clean-registry
clean-registry:
	@echo "Stopping and removing the local Podman registry..."
	podman stop local-registry || true
	podman rm local-registry || true
	@echo "Local registry has been stopped and removed."

# Build the Go binary
.PHONY: build-local
build-local:
	@echo "Building server binary for local testing..."
	go build -o $(BUILD_DIR)/server ./cmd/main.go
	@echo "Build complete! Binary is located at ./server"

# Run the server locally
.PHONY: run-local
run-local: build-local gen-certs
	@echo "Running server locally on port 8443 with generated certificates..."
	IMAGE_NAME=${IMAGE_NAME} \
	IMAGE_REGISTRY=localhost:5000 \
	CA_PUBLIC_KEY=certs/server.crt \
	PRIVATE_KEY=certs/server.key \
	SERVER_PORT=8443 \
	./server

# Build the container image
.PHONY: build-image
build-image:
	@echo "Building container image..."
	podman build -t $(REGISTRY)/vddk-builder:$(TAG) -f Containerfile .

# Push the container image to Quay.io
.PHONY: push-image
push-image:
	@echo "Pushing container image to Quay.io..."
	podman push $(REGISTRY)/vddk-builder:$(TAG)

# Deploy the Pod and Service to the OpenShift cluster
.PHONY: deploy
deploy:
	@echo "Deploying to OpenShift cluster..."
	oc apply -f $(DEPLOY_YAML)

# Clean up OpenShift resources
.PHONY: clean
clean:
	@echo "Cleaning up OpenShift resources..."
	oc delete -f $(DEPLOY_YAML)

# Format the Go code
.PHONY: format
format:
	@echo "Formatting Go code..."
	gofmt -s -w .
	@echo "Go code formatted successfully!"

# Full workflow: build, push, and deploy
.PHONY: all
all: build-image push-image deploy
	@echo "Server successfully deployed to OpenShift!"

# Help menu
.PHONY: help
help:
	@echo "Available tasks:"
	@echo "  goinit              Initialize the Go project for the first time with required tools"
	@echo "  gen-certs           Generate self-signed certificates for local testing"
	@echo "  run-registry        Start a local Podman registry on port 5000 without authentication"
	@echo "  clean-registry      Stop and remove the local Podman registry"
	@echo "  build-local         Build the server binary for local testing"
	@echo "  run-local           Run the server locally on port 8443 with generated certificates"
	@echo "  build-image         Build the container image"
	@echo "  push-image          Push the container image to Quay.io"
	@echo "  deploy              Deploy the server to the OpenShift cluster"
	@echo "  clean               Remove the deployed OpenShift resources"
	@echo "  format              Format the Go code"
	@echo "  all                 Build, push, and deploy the server"
