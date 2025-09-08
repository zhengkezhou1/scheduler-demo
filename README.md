# Kubernetes Scheduler Webhook Demo

This project demonstrates a Kubernetes mutating admission webhook that intelligently schedules pods based on workload characteristics, implementing cost-optimized node affinity and topology spread constraints.

## 🎯 Core Features

### Smart Node Affinity
- **Stateless Applications** (Deployments): Automatically prefer **SPOT nodes** for cost optimization
- **Stateful Applications** (StatefulSets): Automatically prefer **ON-DEMAND nodes** for stability
- Dynamic detection based on pod's OwnerReference chain

### Topology Spread Constraints
- Ensures balanced pod distribution across different node capacity types
- Dynamic MaxSkew calculation based on cluster topology
- Prevents scheduling hotspots and improves fault tolerance

### Intelligent Scheduling Logic
- Webhook intercepts pod creation during admission control
- Analyzes workload type (Deployment vs StatefulSet) 
- Applies appropriate scheduling policies automatically

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   API Server    │───▶│  Webhook Server  │───▶│  Scheduled Pod  │
│                 │    │                  │    │                 │
│ Pod Creation    │    │ • Node Affinity  │    │ • spot/on-demand│
│ Request         │    │ • Topology TSC   │    │ • Balanced      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Docker
- kubectl  
- kind (Kubernetes in Docker)

### Deploy Everything

```bash
make deploy
```

This will:
1. Create a 7-node kind cluster (2 on-demand + 5 spot nodes)
2. Build and load the webhook Docker image
3. Deploy RBAC, TLS certificates, and webhook server
4. Apply webhook configuration
5. Deploy test applications (stateless + stateful)

### Check Scheduling Results

View overall status:
```bash
make status
```

Focused distribution analysis:
```bash
make distribution
```

## 📊 Example Output

```
🎯 Scheduler Demo - Node Affinity Distribution Analysis
================================================================

📋 Webhook Strategy:
  • Stateless workloads (Deployments) → prefer SPOT nodes (cost optimization)
  • Stateful workloads (StatefulSets) → prefer ON-DEMAND nodes (stability)

📊 Current Distribution:

  🔄 batch-worker (Deployment/Stateless):
    15 pods on kind-worker3 (spot)
    12 pods on kind-worker4 (spot)
    8 pods on kind-worker5 (spot)

  🗄️  mock-database (StatefulSet/Stateful):  
    25 pods on kind-worker (on-demand)
    23 pods on kind-worker2 (on-demand)
    2 pods on kind-worker5 (spot)
```

## 🔧 Configuration

### Node Types

The demo cluster simulates two node capacity types:
- **on-demand**: Stable, higher-cost nodes (kind-worker, kind-worker2)
- **spot**: Cost-effective, preemptible nodes (kind-worker3-7)

## 🧪 Testing Scenarios

### Scale Stateless Applications
```bash
kubectl scale deployment batch-worker --replicas=30
make distribution
```

### Scale Stateful Applications  
```bash
kubectl scale statefulset mock-database --replicas=20
make distribution
```

### Observe Scheduling Behavior
- Stateless pods should concentrate on spot nodes
- Stateful pods should prefer on-demand nodes
- Distribution should respect topology spread constraints

## 🧹 Cleanup

```bash
make clean
kind delete cluster
```

## 📚 Implementation Details

### Files Structure
- `pkg/webhook/pod.go`: Pod admission logic and scheduling policies
- `pkg/kube/`: Kubernetes client utilities
- `webhook-config.yaml`: MutatingWebhookConfiguration
- `rbac.yaml`: Service account and permissions
