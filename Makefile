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
deploy-test-workeloads:
	@echo "🧪 Deploying test workeloads..."
	kubectl apply -f test-deployment.yaml
	kubectl apply -f test-statefulset.yaml
	@echo "✅ Test workeloads deployment completed"

.PHONY: deploy
deploy: create-cluster build-image deploy-rbac create-certs deploy-webhook deploy-test-workeloads
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
	@echo "🧪 Application Pod Status:"
	@kubectl get pods -l 'app in (batch-worker,mock-database)' -o wide
	@echo ""
	@echo "📈 Node Distribution Analysis:"
	@echo ""
	@echo "  🔄 Stateless Apps (Deployments - should prefer SPOT nodes):"
	@echo "    📊 batch-worker distribution:"
	@kubectl get pods -l 'app=batch-worker' -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' | sort | uniq -c | sed 's/^/      /'
	@echo ""
	@echo "  🗄️  Stateful Apps (StatefulSets - should prefer ON-DEMAND nodes):"
	@echo "    📊 mock-database distribution:"
	@kubectl get pods -l 'app=mock-database' -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' | sort | uniq -c | sed 's/^/      /'
	@echo ""
	@echo "  🎯 Node Type Summary:"
	@echo "    SPOT nodes (cost-optimized for stateless):"
	@kubectl get nodes -l node.kubernetes.io/capacity=spot --no-headers | awk '{print "      " $$1}'
	@echo "    ON-DEMAND nodes (stable for stateful):"
	@kubectl get nodes -l node.kubernetes.io/capacity=on-demand --no-headers | awk '{print "      " $$1}'

# Distribution analysis - focused view for demo purposes
.PHONY: distribution
distribution:
	@echo "🎯 Scheduler Demo - Node Affinity Distribution Analysis"
	@echo "================================================================"
	@echo ""
	@echo "📋 Webhook Strategy:"
	@echo "  • Stateless workloads (Deployments) → prefer SPOT nodes (cost optimization)"
	@echo "  • Stateful workloads (StatefulSets) → prefer ON-DEMAND nodes (stability)"
	@echo ""
	@echo "📊 Current Distribution:"
	@echo ""
	@echo "  🔄 batch-worker (Deployment/Stateless):"
	@batch_count=$$(kubectl get pods -l app=batch-worker --no-headers 2>/dev/null | wc -l | tr -d ' '); \
	if [ $$batch_count -gt 0 ]; then \
		kubectl get pods -l app=batch-worker -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' | sort | uniq -c | while read count node; do \
			node_type=$$(kubectl get node $$node -o jsonpath='{.metadata.labels.node\.kubernetes\.io/capacity}' 2>/dev/null || echo "unknown"); \
			echo "    $$count pods on $$node ($$node_type)"; \
		done; \
	else \
		echo "    No batch-worker pods found"; \
	fi
	@echo ""
	@echo "  🗄️  mock-database (StatefulSet/Stateful):"
	@db_count=$$(kubectl get pods -l app=mock-database --no-headers 2>/dev/null | wc -l | tr -d ' '); \
	if [ $$db_count -gt 0 ]; then \
		kubectl get pods -l app=mock-database -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' | sort | uniq -c | while read count node; do \
			node_type=$$(kubectl get node $$node -o jsonpath='{.metadata.labels.node\.kubernetes\.io/capacity}' 2>/dev/null || echo "unknown"); \
			echo "    $$count pods on $$node ($$node_type)"; \
		done; \
	else \
		echo "    No mock-database pods found"; \
	fi
	@echo ""
	@echo "💡 Expected behavior: batch-worker pods should mostly be on SPOT nodes,"
	@echo "   while mock-database pods should be on ON-DEMAND nodes."