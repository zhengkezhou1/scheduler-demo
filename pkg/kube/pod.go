package kube

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type podClient struct {
	v1.PodInterface
}

func NewPodClient(namespace string) podClient {
	client := podClient{}
	client.PodInterface = KubeClientset().CoreV1().Pods(namespace)
	return client
}

func (p *podClient) CreatePod(pod *apiv1.Pod) (*apiv1.Pod, error) {
	fmt.Println("Creating pod...")
	return p.Create(context.TODO(), pod, metav1.CreateOptions{})
}
