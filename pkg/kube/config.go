package kube

import (
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func defaultConfig() *rest.Config {
	// First try in-cluster config (for pods running in Kubernetes)
	config, err := rest.InClusterConfig()
	if err == nil {
		return config
	}

	// Fall back to kubeconfig file
	var kubeconfigPath string
	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// use the current context in kubeconfig
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func KubeClientset() *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(defaultConfig())
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
