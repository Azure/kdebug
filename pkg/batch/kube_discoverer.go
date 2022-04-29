package batch

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubeBatchDiscoverer struct {
	client *kubernetes.Clientset
}

func NewKubeBatchDiscoverer(client *kubernetes.Clientset) *KubeBatchDiscoverer {
	return &KubeBatchDiscoverer{
		client: client,
	}
}

func (d *KubeBatchDiscoverer) Discover() ([]string, error) {
	if d.client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := d.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Fail to list nodes from API server: %+v", err)
	}

	var names []string
	for _, node := range resp.Items {
		names = append(names, node.ObjectMeta.Name)
	}

	return names, nil
}
