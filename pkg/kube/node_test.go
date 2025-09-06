package kube

import "testing"

func TestNode(t *testing.T) {
	client := NewNodeClient()
	schedulerNodes, err := client.ListSchedulerNodes()
	if err != nil {
		t.Errorf("Failed to list scheduler nodes: %v", err)
	}

	if schedulerNodes == nil {
		t.Errorf("Scheduler nodes is nil")
	}

	if len(schedulerNodes.SpotNodes) != 5 {
		t.Errorf("expected 5 spot nodes, got %d", len(schedulerNodes.SpotNodes))
	}

	if len(schedulerNodes.OnDemandNodes) != 2 {
		t.Errorf("expected 2 on-demand nodes, got %d", len(schedulerNodes.OnDemandNodes))
	}
}
