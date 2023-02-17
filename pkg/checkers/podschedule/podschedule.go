package podschedule

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/kdebug/pkg/base"
)

type PodScheduleChecker struct {
}

func New() *PodScheduleChecker {
	return &PodScheduleChecker{}
}

func (c *PodScheduleChecker) Name() string {
	return "PodSchedule"
}

func (c *PodScheduleChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	results := []*base.CheckResult{}

	if ctx.KubeClient != nil {
		results = append(results, c.checkPodSchedule(ctx.KubeClient)...)
	} else {
		log.Debugf("Skip %s due to missing Kubernetes config", c.Name())
	}

	return results, nil
}

func (c *PodScheduleChecker) checkPodSchedule(clientset *kubernetes.Clientset) []*base.CheckResult {
	results := []*base.CheckResult{}

	// List all pods
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Fail to list pods")
		return results
	}

	// Group pods by replicaset
	podsByRs := make(map[string][]corev1.Pod)
	for _, pod := range pods.Items {
		if pod.ObjectMeta.OwnerReferences == nil || len(pod.ObjectMeta.OwnerReferences) == 0 {
			continue
		}

		ownerRef := pod.ObjectMeta.OwnerReferences[0]
		if ownerRef.APIVersion == "apps/v1" &&
			ownerRef.Kind == "ReplicaSet" {

			rsName := pod.ObjectMeta.Namespace + "/" + ownerRef.Name
			if rsPods, ok := podsByRs[rsName]; ok {
				podsByRs[rsName] = append(rsPods, pod)
			} else {
				podsByRs[rsName] = []corev1.Pod{pod}
			}
		}
	}

	// Check replica sets
	for rsName, rsPods := range podsByRs {
		if len(rsPods) <= 1 {
			continue
		}

		results = append(results, c.checkPodsScheduleInReplicaSet(rsName, rsPods))
	}

	return results
}

func (c *PodScheduleChecker) checkPodsScheduleInReplicaSet(rsName string, pods []corev1.Pod) *base.CheckResult {
	if len(pods) <= 1 {
		panic("Should not be called with less than 2 pods")
	}

	node := ""
	for _, pod := range pods {
		if node == "" {
			node = pod.Spec.NodeName
		} else if node != pod.Spec.NodeName {
			return &base.CheckResult{
				Checker:     c.Name(),
				Description: fmt.Sprintf("Pods in replica set %s are scheduled to different nodes", rsName),
			}
		}
	}
	return &base.CheckResult{
		Checker: c.Name(),
		Error:   fmt.Sprintf("All pods of replica set %s are scheduled on same node", rsName),
		Recommendations: []string{
			"Please reference to document to set Affinity and anti-affinity: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity",
		},
	}
}
