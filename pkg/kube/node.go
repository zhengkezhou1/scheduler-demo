package kube

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	NodeCapacityLabel = "node.kubernetes.io/capacity"

	NodeCapacityTypeSpot     string = "spot"
	NodeCapacityTypeOnDemand string = "on-demand"
)

type NodeClient struct {
	v1.NodeInterface
}

type SchedulerNode struct {
	SpotNodes     []apiv1.Node
	OnDemandNodes []apiv1.Node
}

func NewNodeClient() *NodeClient {
	client := NodeClient{
		NodeInterface: KubeClientset().CoreV1().Nodes(),
	}
	return &client
}

func (c *NodeClient) ListSchedulerNodes() (*SchedulerNode, error) {
	schedulerNode := SchedulerNode{
		SpotNodes:     make([]apiv1.Node, 0),
		OnDemandNodes: make([]apiv1.Node, 0),
	}
	allNodes, err := c.NodeInterface.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range allNodes.Items {
		if node.Labels[NodeCapacityLabel] == NodeCapacityTypeSpot {
			schedulerNode.SpotNodes = append(schedulerNode.SpotNodes, node)
		}
		if node.Labels[NodeCapacityLabel] == NodeCapacityTypeOnDemand {
			schedulerNode.OnDemandNodes = append(schedulerNode.OnDemandNodes, node)
		}
	}

	return &schedulerNode, nil
}
