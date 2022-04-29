package batch

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestMatchNode(t *testing.T) {
	d := &KubeBatchDiscoverer{}
	node := &corev1.Node{}
	if !d.matchNode(node) {
		t.Errorf("Expect matchNode == true when not specifying unready but got false")
	}

	d = &KubeBatchDiscoverer{unready: true}
	node = &corev1.Node{
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}
	if !d.matchNode(node) {
		t.Errorf("Expect matchNode == true when specifying unready and node is unready but got false")
	}

	node = &corev1.Node{
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	if d.matchNode(node) {
		t.Errorf("Expect matchNode == false when specifying unready and node is ready but got true")
	}
}
