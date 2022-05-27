package batch

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubeBatchDiscoverer struct {
	client        *kubernetes.Clientset
	labelSelector string
	unready       bool
}

func NewKubeBatchDiscoverer(client *kubernetes.Clientset, labelSelector string, unready bool) *KubeBatchDiscoverer {
	return &KubeBatchDiscoverer{
		client:        client,
		labelSelector: labelSelector,
		unready:       unready,
	}
}

func (d *KubeBatchDiscoverer) Discover() ([]string, error) {
	if d.client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := d.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: d.labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("Fail to list nodes from API server: %+v", err)
	}

	var names []string
	for _, node := range resp.Items {
		if d.matchNode(&node) {
			names = append(names, node.ObjectMeta.Name)
		}
	}

	return names, nil
}

func (d *KubeBatchDiscoverer) matchNode(node *corev1.Node) bool {
	if d.unready {
		// Unready only
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				return cond.Status != corev1.ConditionTrue
			}
		}
	}

	return true
}
