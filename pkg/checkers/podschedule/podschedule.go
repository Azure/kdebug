package dns

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/dustin/go-humanize"
)

const (
	WarnSizeThreshold = 800 * (1 << 10) // 800 KB
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

	podList, err := clientset.CoreV1().Pods("namespace").List(context.Background(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Fail to list pod under rs")
	}

}

func (c *PodScheduleChecker) checkObjectSize(kind, ns, name string, obj interface{}) *base.CheckResult {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil
	}

	if len(data) > WarnSizeThreshold {
		return &base.CheckResult{
			Checker:     c.Name(),
			Error:       fmt.Sprintf("%s %s/%s reaching size limit.", kind, ns, name),
			Description: fmt.Sprintf("%s %s/%s of size %s is reaching size limit. It cannot exceed 1MiB.", kind, ns, name, humanize.Bytes(uint64(len(data)))),
			Recommendations: []string{
				"Consider mounting a volume or use a separate database or file service.",
			},
		}
	}

	return &base.CheckResult{
		Checker:     c.Name(),
		Description: fmt.Sprintf("%s %s/%s of size %s is not reaching size limit.", kind, ns, name, humanize.Bytes(uint64(len(data)))),
	}
}
