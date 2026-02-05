.PHONY: build run test docker-build docker-run k8s-deploy k8s-delete compose-up compose-down clean help

# Variables
APP_NAME=bank-api
DOCKER_IMAGE=$(APP_NAME):latest
NAMESPACE=banking-system

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go application
	@echo "Building $(APP_NAME)..."
	go build -o $(APP_NAME) ./cmd/server

run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	go run ./cmd/server

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

tidy: ## Run go mod tidy
	@echo "Tidying dependencies..."
	go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker image locally
	@echo "Running Docker container..."
	docker run -p 8080:8080 --rm --name $(APP_NAME) $(DOCKER_IMAGE)


k8s-deploy: docker-build ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	@./scripts/deploy-k8s.sh

k8s-delete: ## Delete Kubernetes deployment
	@echo "Deleting Kubernetes deployment..."
	kubectl delete namespace $(NAMESPACE)

k8s-status: ## Check Kubernetes deployment status
	@echo "Checking Kubernetes status..."
	kubectl get all -n $(NAMESPACE)

k8s-logs: ## Show Kubernetes logs
	kubectl logs -n $(NAMESPACE) -l app=$(APP_NAME) -f

test-api: ## Test API endpoints
	@./scripts/test-api.sh

test-api-k8s: ## Test API endpoints on Kubernetes
	@./scripts/test-api.sh http://localhost:30080

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	rm -rf data/
	go clean

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

.DEFAULT_GOAL := help
