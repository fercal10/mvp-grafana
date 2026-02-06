.PHONY: build run test build-accounts build-transfers build-all k8s-deploy k8s-delete k8s-logs-accounts k8s-logs-transfers clean help

# Variables
ACCOUNTS_APP=accounts-api
TRANSFERS_APP=transfers-api
ACCOUNTS_IMAGE=$(ACCOUNTS_APP):latest
TRANSFERS_IMAGE=$(TRANSFERS_APP):latest
NAMESPACE=banking-system

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: build-all ## Build both microservices (alias of build-all)

build-accounts: ## Build accounts-api binary
	@echo "Building $(ACCOUNTS_APP)..."
	go build -o $(ACCOUNTS_APP) ./cmd/accounts-api

build-transfers: ## Build transfers-api binary
	@echo "Building $(TRANSFERS_APP)..."
	go build -o $(TRANSFERS_APP) ./cmd/transfers-api

build-all: build-accounts build-transfers ## Build both microservices

run: ## Run both microservices locally (run in two terminals: PORT=8080 go run ./cmd/accounts-api and PORT=8081 go run ./cmd/transfers-api)
	@echo "Run accounts-api:  PORT=8080 go run ./cmd/accounts-api"
	@echo "Run transfers-api: PORT=8081 go run ./cmd/transfers-api"

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

tidy: ## Run go mod tidy
	@echo "Tidying dependencies..."
	go mod tidy

docker-build-accounts: ## Build accounts-api Docker image
	@echo "Building Docker image $(ACCOUNTS_IMAGE)..."
	docker build -f Dockerfile.accounts -t $(ACCOUNTS_IMAGE) .

docker-build-transfers: ## Build transfers-api Docker image
	@echo "Building Docker image $(TRANSFERS_IMAGE)..."
	docker build -f Dockerfile.transfers -t $(TRANSFERS_IMAGE) .

docker-build-microservices: docker-build-accounts docker-build-transfers ## Build both microservice images

k8s-deploy: ## Deploy to Kubernetes (script builds both images and applies manifests)
	@echo "Deploying to Kubernetes..."
	@./scripts/deploy-k8s.sh

k8s-delete: ## Delete Kubernetes deployment
	@echo "Deleting Kubernetes deployment..."
	kubectl delete namespace $(NAMESPACE)

k8s-status: ## Check Kubernetes deployment status
	@echo "Checking Kubernetes status..."
	kubectl get all -n $(NAMESPACE)

k8s-logs-accounts: ## Show Kubernetes logs for accounts-api
	kubectl logs -n $(NAMESPACE) -l app=$(ACCOUNTS_APP) -f

k8s-logs-transfers: ## Show Kubernetes logs for transfers-api
	kubectl logs -n $(NAMESPACE) -l app=$(TRANSFERS_APP) -f

test-api: ## Test API endpoints
	@./scripts/test-api.sh

test-api-k8s: ## Test API endpoints on Kubernetes (accounts + transfers)
	@./scripts/test-api.sh http://localhost:30080 http://localhost:30081

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(ACCOUNTS_APP) $(TRANSFERS_APP)
	rm -rf data/
	go clean

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

.DEFAULT_GOAL := help
