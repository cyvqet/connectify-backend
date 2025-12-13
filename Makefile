# ===== Colors =====
GREEN  := \033[32m
YELLOW := \033[33m
BLUE   := \033[34m
RED    := \033[31m
CYAN   := \033[36m
BOLD   := \033[1m
RESET  := \033[0m

# ===== Log helpers =====
INFO    = echo "$(BLUE)[INFO]$(RESET)"
SUCCESS = echo "$(GREEN)[SUCCESS]$(RESET)"
WARN    = echo "$(YELLOW)[WARN]$(RESET)"
ERROR   = echo "$(RED)[ERROR]$(RESET)"
STEP    = echo "$(CYAN)==>$(RESET)"

# Define phony targets (not real files, always execute commands)
.PHONY: docker docker-clean docker-deploy mysql-deploy mysql-clean redis-deploy redis-clean ingress-deploy ingress-clean redeploy all
.PHONY: dev dev-up dev-down dev-run dev-clean help k8s k8s-clean

# Docker image build - compile Go program and build Docker image
docker:
	@$(STEP) "$(BOLD)Step 1$(RESET): Removing old connectify binary..."
	@rm -f connectify || true

	@$(INFO) "Tidying Go module dependencies..."
	@go mod tidy

	@$(STEP) "$(BOLD)Step 2$(RESET): Cross-compiling Go program (Linux ARM)..."
	@GOOS=linux GOARCH=arm go build -tags=k8s -o connectify .
	# GOOS=linux: target operating system
	# GOARCH=arm: target CPU architecture
	# -tags=k8s: enable k8s build tag (conditional compilation)
	# -o connectify: output binary name

	@$(INFO) "Removing old Docker image (if exists)..."
	@docker rmi -f cyvqet/connectify:v1.0 2>/dev/null || true

	@$(STEP) "$(BOLD)Step 3$(RESET): Building new Docker image..."
	@docker build -t cyvqet/connectify:v1.0 .
	# -t: tag the image (username/image:version)
	# . : use current directory as build context (must contain Dockerfile)

	@$(INFO) "Cleaning dangling Docker images..."
	@docker image prune -f

	@$(INFO) "Cleaning build artifacts..."
	@rm -f connectify

	@$(SUCCESS) "Docker image build completed"

# Clean Kubernetes resources - remove deployed services and deployments
docker-clean:
	@$(INFO) "Cleaning Kubernetes resources..."
	@kubectl delete service connectify-record 2>/dev/null || true
	@kubectl delete deployment connectify-record-service 2>/dev/null || true
	@$(INFO) "Waiting for cleanup to complete..."
	@sleep 2
	@$(SUCCESS) "Kubernetes resources cleanup completed"

# Deploy application to Kubernetes
docker-deploy:
	@$(INFO) "Deploying application to Kubernetes..."
	@kubectl apply -f deploy/k8s/connectify-deployment.yaml
	@kubectl apply -f deploy/k8s/connectify-service.yaml

	@$(INFO) "Waiting for Pods to start..."
	@sleep 5
	@kubectl get pods -l app=connectify-record

	@$(SUCCESS) "Application deployed successfully"

# Deploy MySQL to Kubernetes - create PV, PVC, Deployment, and Service
mysql-deploy:
	@$(INFO) "Deploying MySQL to Kubernetes..."
	@kubectl apply -f deploy/k8s/connectify-mysql-pv.yaml
	@kubectl apply -f deploy/k8s/connectify-mysql-pvc.yaml
	@kubectl apply -f deploy/k8s/connectify-mysql-deployment.yaml
	@kubectl apply -f deploy/k8s/connectify-mysql-service.yaml

	@$(INFO) "Waiting for MySQL Pod to start..."
	@sleep 10
	@kubectl get pods -l app=connectify-record-mysql > /dev/null || true
	@kubectl get pv,pvc > /dev/null || true

	@$(INFO) "Creating database: connectify..."
	@kubectl exec -it $$(kubectl get pods -l app=connectify-record-mysql -o jsonpath='{.items[0].metadata.name}') -- \
		mysql -uroot -proot -e "CREATE DATABASE IF NOT EXISTS connectify;"

	@$(SUCCESS) "MySQL deployed and database ready"

# Clean MySQL Kubernetes resources
mysql-clean:
	@$(INFO) "Cleaning MySQL Kubernetes resources..."
	@kubectl delete service connectify-record-mysql 2>/dev/null || true
	@kubectl delete deployment connectify-record-mysql 2>/dev/null || true
	@kubectl delete pvc connectify-mysql-pvc 2>/dev/null || true
	@kubectl delete pv connectify-mysql-pv 2>/dev/null || true
	@$(INFO) "Waiting for cleanup to complete..."
	@sleep 2
	@$(SUCCESS) "MySQL cleanup completed"

# Deploy Redis to Kubernetes
redis-deploy:
	@$(INFO) "Deploying Redis to Kubernetes..."
	@kubectl apply -f deploy/k8s/connectify-redis-deployment.yaml
	@kubectl apply -f deploy/k8s/connectify-redis-service.yaml

	@$(INFO) "Waiting for Redis Pod to start..."
	@sleep 5
	@kubectl get pods -l app=connectify-record-redis

	@$(SUCCESS) "Redis deployed successfully"

# Clean Redis Kubernetes resources
redis-clean:
	@$(INFO) "Cleaning Redis Kubernetes resources..."
	@kubectl delete service connectify-record-redis 2>/dev/null || true
	@kubectl delete deployment connectify-record-redis 2>/dev/null || true
	@$(INFO) "Waiting for cleanup to complete..."
	@sleep 2
	@$(SUCCESS) "Redis cleanup completed"

# Deploy Ingress to Kubernetes
ingress-deploy:
	@$(INFO) "Deploying Ingress to Kubernetes..."
	@kubectl apply -f deploy/k8s/connectify-ingress.yaml
	@$(SUCCESS) "Ingress deployment completed"
	@$(INFO) "Access URL: http://localhost/"

# Clean Ingress resources
ingress-clean:
	@$(INFO) "Cleaning Ingress resources..."
	@kubectl delete ingress connectify-record-ingress 2>/dev/null || true
	@$(SUCCESS) "Ingress cleanup completed"

# Rebuild image and perform rolling update in Kubernetes
redeploy:
	@$(INFO) "Building new Docker image..."
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -tags=k8s -o connectify .
	@docker build -t cyvqet/connectify:v1.0 .

	@$(WARN) "Pushing image is commented out (enable if needed)"
	# @docker push cyvqet/connectify:v1.0

	@$(INFO) "Restarting Deployment (rolling update)..."
	@kubectl rollout restart deployment connectify-record-service

	@$(INFO) "Cleaning build artifacts..."
	@rm -f connectify

	@$(SUCCESS) "Rolling update completed successfully"

# Clean all Kubernetes resources
k8s-clean:
	@$(WARN) "Cleaning ALL Kubernetes resources..."
	@kubectl delete service connectify-record 2>/dev/null || true
	@kubectl delete deployment connectify-record-service 2>/dev/null || true
	@kubectl delete service connectify-record-mysql 2>/dev/null || true
	@kubectl delete deployment connectify-record-mysql 2>/dev/null || true
	@kubectl delete pvc connectify-mysql-pvc 2>/dev/null || true
	@kubectl delete pv connectify-mysql-pv 2>/dev/null || true
	@kubectl delete service connectify-record-redis 2>/dev/null || true
	@kubectl delete deployment connectify-record-redis 2>/dev/null || true
	@kubectl delete ingress connectify-record-ingress 2>/dev/null || true
	@sleep 2
	@$(SUCCESS) "All Kubernetes resources cleaned"

# One-click Kubernetes deployment (stops local dev first)
k8s: 
	@echo ""
	@echo "$(YELLOW)==========================================$(RESET)"
	@echo "$(YELLOW)  Switching to Kubernetes Environment$(RESET)"
	@echo "$(YELLOW)==========================================$(RESET)"
	@echo ""
	@$(WARN) "Stopping local development environment first..."
	@cd deploy/docker && docker-compose down 2>/dev/null || true
	@$(WARN) "Killing non-Docker processes on conflicting ports..."
	@lsof -i :8080 2>/dev/null | grep -v 'com.docke' | awk 'NR>1 {print $$2}' | xargs kill -9 2>/dev/null || true
	@lsof -i :6379 2>/dev/null | grep -v 'com.docke' | awk 'NR>1 {print $$2}' | xargs kill -9 2>/dev/null || true
	@lsof -i :13316 2>/dev/null | grep -v 'com.docke' | awk 'NR>1 {print $$2}' | xargs kill -9 2>/dev/null || true
	@sleep 2
	@$(MAKE) docker docker-clean mysql-deploy redis-deploy docker-deploy ingress-deploy
	@echo ""
	@echo "$(GREEN)==========================================$(RESET)"
	@echo "$(GREEN)  K8s Deployment completed!$(RESET)"
	@echo "$(GREEN)==========================================$(RESET)"
	@echo ""
	@$(INFO) "Access URL: http://localhost/test"
	@echo ""
	@$(INFO) "Common commands:"
	@echo "  kubectl get pods"
	@echo "  kubectl logs -f deployment/connectify-record-service"
	@echo "  make k8s-clean  (clean all K8s resources)"
	@echo ""

# Legacy: One-click build, clean, and deploy - full CI/CD workflow
all: docker docker-clean mysql-deploy redis-deploy docker-deploy ingress-deploy
	@echo ""
	@echo "$(GREEN)==========================================$(RESET)"
	@echo "$(GREEN)  Deployment completed successfully!$(RESET)"
	@echo "$(GREEN)==========================================$(RESET)"
	@echo ""
	@$(INFO) "Access URL: http://localhost/test"
	@echo ""
	@$(INFO) "Common commands:"
	@echo "  kubectl get pods"
	@echo "  kubectl logs -f deployment/connectify-record-service"
	@echo "  kubectl logs -f -l app=connectify-record --all-containers"
	@echo "  kubectl rollout restart deployment/connectify-record-service"
	@echo "  kubectl exec -it \$$(kubectl get pods -l app=connectify-record -o jsonpath='{.items[0].metadata.name}') -- sh"
	@echo "  make docker-clean mysql-clean redis-clean ingress-clean"
	@echo ""

# ==========================================
# Local Development Commands
# ==========================================

# Start local dependencies (MySQL + Redis) - stops K8s first
dev-up:
	@echo ""
	@echo "$(YELLOW)==========================================$(RESET)"
	@echo "$(YELLOW)  Switching to Local Dev Environment$(RESET)"
	@echo "$(YELLOW)==========================================$(RESET)"
	@echo ""
	@$(WARN) "Stopping ALL Kubernetes services first (if running)..."
	@kubectl delete service connectify-record 2>/dev/null || true
	@kubectl delete deployment connectify-record-service 2>/dev/null || true
	@kubectl delete ingress connectify-record-ingress 2>/dev/null || true
	@kubectl delete service connectify-record-mysql 2>/dev/null || true
	@kubectl delete deployment connectify-record-mysql 2>/dev/null || true
	@kubectl delete service connectify-record-redis 2>/dev/null || true
	@kubectl delete deployment connectify-record-redis 2>/dev/null || true
	@$(WARN) "Killing non-Docker processes on conflicting ports (8080, 6379)..."
	@lsof -i :8080 2>/dev/null | grep -v 'com.docke' | awk 'NR>1 {print $$2}' | xargs kill -9 2>/dev/null || true
	@lsof -i :6379 2>/dev/null | grep -v 'com.docke' | awk 'NR>1 {print $$2}' | xargs kill -9 2>/dev/null || true
	@sleep 2
	@$(INFO) "Starting local development dependencies..."
	@cd deploy/docker && docker-compose up -d
	@$(INFO) "Waiting for MySQL to be ready..."
	@sleep 5
	@$(SUCCESS) "Local dependencies started"
	@echo ""
	@$(INFO) "MySQL: localhost:13316 (user: root, password: root)"
	@$(INFO) "Redis: localhost:6379"

# Stop local dependencies
dev-down:
	@$(INFO) "Stopping local development dependencies..."
	@cd deploy/docker && docker-compose down
	@$(SUCCESS) "Local dependencies stopped"

# Clean local dependencies (including volumes)
dev-clean:
	@$(INFO) "Cleaning local development environment..."
	@cd deploy/docker && docker-compose down -v
	@$(SUCCESS) "Local environment cleaned (volumes removed)"

# Run the application locally
dev-run:
	@$(INFO) "Running application locally..."
	@go run .

# One-click local development - start deps and run app
dev: dev-up
	@echo ""
	@echo "$(GREEN)==========================================$(RESET)"
	@echo "$(GREEN)  Local Development Environment Ready$(RESET)"
	@echo "$(GREEN)==========================================$(RESET)"
	@echo ""
	@$(INFO) "Starting application on :8080..."
	@go run .

.PHONY: logs k8s-logs k8s-status k8s-shell mysql-shell redis-cli status

# Show help
help:
	@echo ""
	@echo "$(BOLD)Connectify Backend - Available Commands$(RESET)"
	@echo ""
	@echo "$(CYAN)Quick Switch:$(RESET)"
	@echo "  make dev          - Switch to local development"
	@echo "  make k8s          - Switch to Kubernetes"
	@echo ""
	@echo "$(CYAN)Local Development:$(RESET)"
	@echo "  make dev          - One-click: stop K8s + start deps + run app"
	@echo "  make dev-up       - Start MySQL & Redis containers only"
	@echo "  make dev-down     - Stop containers"
	@echo "  make dev-run      - Run app only (deps already running)"
	@echo "  make dev-clean    - Stop containers & remove volumes"
	@echo ""
	@echo "$(CYAN)Kubernetes:$(RESET)"
	@echo "  make k8s          - One-click: stop local + full K8s deploy"
	@echo "  make k8s-status   - Show pods, services, ingress status"
	@echo "  make k8s-logs     - Tail application logs"
	@echo "  make k8s-shell    - Shell into app container"
	@echo "  make k8s-clean    - Clean all K8s resources"
	@echo "  make redeploy     - Rebuild & rolling update"
	@echo ""
	@echo "$(CYAN)Database & Cache:$(RESET)"
	@echo "  make mysql-shell  - Connect to MySQL CLI"
	@echo "  make redis-cli    - Connect to Redis CLI"
	@echo ""
	@echo "$(CYAN)Debugging:$(RESET)"
	@echo "  make status       - Show current environment status"
	@echo "  make logs         - Tail logs (auto-detect environment)"
	@echo ""

# ==========================================
# Status & Logs Commands
# ==========================================

# Show current environment status
status:
	@echo ""
	@echo "$(BOLD)Current Environment Status$(RESET)"
	@echo ""
	@echo "$(CYAN)Port 8080:$(RESET)"
	@lsof -i :8080 2>/dev/null || echo "  (not in use)"
	@echo ""
	@echo "$(CYAN)Docker Compose:$(RESET)"
	@cd deploy/docker && docker-compose ps 2>/dev/null || echo "  (not running)"
	@echo ""
	@echo "$(CYAN)Kubernetes:$(RESET)"
	@kubectl get pods -l app=connectify-record 2>/dev/null || echo "  (no pods)"
	@echo ""

# Tail logs (auto-detect environment)
logs:
	@if kubectl get pods -l app=connectify-record 2>/dev/null | grep -q Running; then \
		$(INFO) "Tailing K8s logs..."; \
		kubectl logs -f -l app=connectify-record --all-containers; \
	else \
		$(INFO) "K8s not running. Use 'make dev' to start local development."; \
	fi

# K8s: Show status of all resources
k8s-status:
	@echo ""
	@echo "$(BOLD)Kubernetes Resources$(RESET)"
	@echo ""
	@echo "$(CYAN)Pods:$(RESET)"
	@kubectl get pods -l app=connectify-record 2>/dev/null || true
	@kubectl get pods -l app=connectify-record-mysql 2>/dev/null || true
	@kubectl get pods -l app=connectify-record-redis 2>/dev/null || true
	@echo ""
	@echo "$(CYAN)Services:$(RESET)"
	@kubectl get svc | grep connectify 2>/dev/null || true
	@echo ""
	@echo "$(CYAN)Ingress:$(RESET)"
	@kubectl get ingress 2>/dev/null || true
	@echo ""

# K8s: Tail application logs
k8s-logs:
	@$(INFO) "Tailing K8s application logs (Ctrl+C to stop)..."
	@kubectl logs -f -l app=connectify-record --all-containers

# K8s: Shell into app container
k8s-shell:
	@$(INFO) "Connecting to app container..."
	@kubectl exec -it $$(kubectl get pods -l app=connectify-record -o jsonpath='{.items[0].metadata.name}') -- sh

# Connect to MySQL CLI
mysql-shell:
	@if cd deploy/docker && docker-compose ps 2>/dev/null | grep -q mysql8; then \
		$(INFO) "Connecting to local MySQL..."; \
		docker exec -it docker-mysql8-1 mysql -uroot -proot connectify; \
	elif kubectl get pods -l app=connectify-record-mysql 2>/dev/null | grep -q Running; then \
		$(INFO) "Connecting to K8s MySQL..."; \
		kubectl exec -it $$(kubectl get pods -l app=connectify-record-mysql -o jsonpath='{.items[0].metadata.name}') -- mysql -uroot -proot connectify; \
	else \
		$(ERROR) "No MySQL found. Run 'make dev' or 'make k8s' first."; \
	fi

# Connect to Redis CLI
redis-cli:
	@if cd deploy/docker && docker-compose ps 2>/dev/null | grep -q redis; then \
		$(INFO) "Connecting to local Redis..."; \
		docker exec -it docker-redis-1 redis-cli; \
	elif kubectl get pods -l app=connectify-record-redis 2>/dev/null | grep -q Running; then \
		$(INFO) "Connecting to K8s Redis..."; \
		kubectl exec -it $$(kubectl get pods -l app=connectify-record-redis -o jsonpath='{.items[0].metadata.name}') -- redis-cli; \
	else \
		$(ERROR) "No Redis found. Run 'make dev' or 'make k8s' first."; \
	fi
