// pkg/client/client.go
package client

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client wraps Kubernetes clients
type Client struct {
	Clientset     kubernetes.Interface
	DynamicClient dynamic.Interface
}

// NewClient creates a new Kubernetes client
func NewClient(config *rest.Config) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Clientset:     clientset,
		DynamicClient: dynamicClient,
	}, nil
}
