# Webhook Demo Makefile
# Encapsulates deployment steps from README

# Variable definitions
CLUSTER_NAME = kind
IMAGE_NAME = webhook:v1
WEBHOOK_NAMESPACE = default

# Default target
.PHONY: all
all: deploy

# 1. Create Kind cluster
.PHONY: create-cluster
create-cluster:
	@echo "🚀 Creating Kind cluster..."
	kind create cluster --config kind-config.yaml
	@echo "✅ Kind cluster creation completed"

# 2. Build and load Docker image
.PHONY: build-image
build-image:
	@echo "🔨 Building Docker image..."
	docker build -t $(IMAGE_NAME) .
	@echo "📦 Loading image to Kind cluster..."
	kind load docker-image $(IMAGE_NAME)
	@echo "✅ Image build and load completed"

# 3. Deploy RBAC resources
.PHONY: deploy-rbac
deploy-rbac:
	@echo "🔐 Deploying RBAC resources..."
	kubectl apply -f rbac.yaml
	@echo "✅ RBAC resources deployment completed"

# 4. Create TLS certificate Secret
.PHONY: create-certs
create-certs:
	@echo "🔐 Creating TLS certificate Secret..."
	kubectl create secret tls webhook-certs --cert=certs/webhook-cert.pem --key=certs/webhook-key.pem --dry-run=client -o yaml | kubectl apply -f -
	@echo "✅ TLS certificate Secret creation completed"

# 5. Deploy Webhook server
.PHONY: deploy-webhook
deploy-webhook:
	@echo "🚀 Deploying Webhook server..."
	kubectl apply -f webhook-server.yaml
	@echo "⏳ Waiting for webhook to be ready..."
	kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=webhook-server --timeout=60s
	@echo "📋 Applying Webhook configuration..."
	kubectl apply -f webhook-config.yaml
	@echo "✅ Webhook deployment completed"

# 6. Deploy test application
.PHONY: deploy-test
deploy-test-deployment:
	@echo "🧪 Deploying test application..."
	kubectl apply -f test-deployment.yaml
	@echo "✅ Test application deployment completed"

.PHONY: deploy
deploy: create-cluster build-image deploy-rbac create-certs deploy-webhook deploy-test-deployment
	@echo "🎉 Complete deployment process finished!"

# Clean up resources
.PHONY: clean
clean:
	@echo "🧹 Cleaning up test resources..."
	kubectl delete -f test-deployment.yaml --ignore-not-found=true
	kubectl delete -f webhook-config.yaml --ignore-not-found=true
	kubectl delete -f webhook-server.yaml --ignore-not-found=true
	kubectl delete -f rbac.yaml --ignore-not-found=true
	kubectl delete secret webhook-certs --ignore-not-found=true
	@echo "✅ Resource cleanup completed"

# Status check
.PHONY: status
status:
	@echo "📊 Cluster status:"
	@kubectl get nodes --show-labels | grep "node.kubernetes.io/capacity"
	@echo ""
	@echo "🧪 Test application status:"
	@kubectl get pods -l app=batch-worker
	@echo ""
	@echo "📈 Pod distribution statistics:"
	@kubectl get pods -l 'app in (batch-worker)' -o wide | grep -v NAME | awk '{print $$7}' | sort | uniq -c