package podschedule

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
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
	return "PodScheduleChecker"
}

func (c *PodScheduleChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	results := []*base.CheckResult{}

	if ctx.KubeClient != nil {
		results = append(results, c.checkPodSchedule(ctx.KubeClient)...)
	} else {
		log.Warn("Skip PodScheduleChecker due to missing kube client")
	}

	return results, nil
}

func (c *PodScheduleChecker) checkPodSchedule(clientset *kubernetes.Clientset) []*base.CheckResult {
	results := []*base.CheckResult{}

	rsl, err := clientset.AppsV1().ReplicaSets("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Fail to list deployments")
		return results
	}

	for _, rs := range rsl.Items {
		if rs.Status.Replicas <= 1 {
			continue
		}
		result := c.checkPodScheduleInReplicaSet(rs, clientset)
		if result != nil {
			results = append(results, result)
		}
	}

	return results
}

func (c *PodScheduleChecker) checkPodScheduleInReplicaSet(rs v1.ReplicaSet, clientset *kubernetes.Clientset) *base.CheckResult {
	selector, err := metav1.LabelSelectorAsSelector(rs.Spec.Selector)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Fail to get selector of rs")
	}

	podList, err := clientset.CoreV1().Pods(rs.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Fail to list pod under rs")
	}

	return c.checkForPodList(rs.Name, podList)
}

func (c *PodScheduleChecker) checkForPodList(rsName string, podList *corev1.PodList) *base.CheckResult {
	if len(podList.Items) > 1 {
		node := ""
		for _, pod := range podList.Items {
			if node == "" {
				node = pod.Spec.NodeName
			} else if node != pod.Spec.NodeName {
				return &base.CheckResult{
					Checker:     c.Name(),
					Description: fmt.Sprintf("Pod in replica set %s are scheduled in good shape", rsName),
				}
			}
		}
		return &base.CheckResult{
			Checker: c.Name(),
			Error:   fmt.Sprintf("All pods of replica set %s are scheduled in the same node", rsName),
			Recommendations: []string{
				"Please reference to document to set Affinity and anti-affinity: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity",
			},
		}
	}

	return &base.CheckResult{
		Checker:     c.Name(),
		Description: fmt.Sprintf("Pod in replica set %s are scheduled in good shape", rsName),
	}
}
