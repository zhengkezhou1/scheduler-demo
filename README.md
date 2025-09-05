# Kubernetes Admission Webhook Demo

This project demonstrates a Kubernetes mutating admission webhook that automatically applies node affinity and topology spread constraints to pods during creation.

## Features

- **NodeAffinity**: Automatically adds preference for spot nodes for stateless applications
- **TopologySpreadConstraints**: Ensures balanced pod distribution across different node types
- **Dynamic Scheduling**: Webhook intercepts pod creation and applies intelligent scheduling policies

## Quick Start

### Prerequisites

- Docker
- kubectl
- kind (Kubernetes in Docker)

### Deploy Everything

```zsh
make deploy
```

This will:
1. Create a kind cluster
2. Build and load the webhook Docker image
3. Create TLS certificates
4. Deploy the webhook server
5. Deploy test applications

### Check Status

```zsh
make status
```

## TODO

- `MaxSkew` should be dynamically adjusted based on the number of spot and on-demand nodes in the data plane, as well as the workload type
- `NodeAffinity` needs to be adjusted according to the actual deployed applications. Stateless applications should be preferentially scheduled to spot nodes, while stateful applications should be forced to be scheduled to on-demand nodes
- Complete testing in auto-scaling scenarios (HPA, VPA, etc.)