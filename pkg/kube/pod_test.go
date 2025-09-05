package kube

import (
	"context"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPod(t *testing.T) {
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-pod",
			Labels: map[string]string{"app": "web-server"},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:  "web",
					Image: "nginx:1.12",
					Ports: []apiv1.ContainerPort{
						{
							Name:          "http",
							Protocol:      apiv1.ProtocolTCP,
							ContainerPort: 80,
						},
					},
				},
			},
		},
	}

	clinet := NewPodClient(metav1.NamespaceDefault)
	created, err := clinet.CreatePod(pod)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	if created == nil {
		t.Fatalf("failed to create pods")
	}

	time.Sleep(time.Second * 15)

	pods, err := clinet.Get(context.TODO(), "test-pod", metav1.GetOptions{})

	if err != nil {
		t.Fail()
	}

	if pods == nil {
		t.Errorf("pod not exist in cluster!")
	}
}
