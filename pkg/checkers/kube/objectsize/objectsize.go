package dns

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/dustin/go-humanize"
)

const (
	WarnSizeThreshold = 800 * (1 << 10) // 800 KB
)

type KubeObjectSizeChecker struct {
}

func New() *KubeObjectSizeChecker {
	return &KubeObjectSizeChecker{}
}

func (c *KubeObjectSizeChecker) Name() string {
	return "KubeObjectSize"
}

func (c *KubeObjectSizeChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	results := []*base.CheckResult{}

	if ctx.KubeClient != nil {
		results = append(results, c.checkConfigMaps(ctx.KubeClient)...)
		results = append(results, c.checkSecrets(ctx.KubeClient)...)
	} else {
		log.Warn("Skip KubeObjectSizeChecker due to missing kube client")
	}

	return results, nil
}

func (c *KubeObjectSizeChecker) checkConfigMaps(clientset *kubernetes.Clientset) []*base.CheckResult {
	results := []*base.CheckResult{}

	cms, err := clientset.CoreV1().ConfigMaps("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Fail to list config maps")
		return results
	}

	for _, cm := range cms.Items {
		result := c.checkObjectSize("ConfigMap", cm.ObjectMeta.Namespace, cm.ObjectMeta.Name, cm)
		if result != nil {
			results = append(results, result)
		}
	}

	return results
}

func (c *KubeObjectSizeChecker) checkSecrets(clientset *kubernetes.Clientset) []*base.CheckResult {
	results := []*base.CheckResult{}

	cms, err := clientset.CoreV1().Secrets("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Fail to list secrets")
		return results
	}

	for _, cm := range cms.Items {
		result := c.checkObjectSize("Secret", cm.ObjectMeta.Namespace, cm.ObjectMeta.Name, cm)
		if result != nil {
			results = append(results, result)
		}
	}

	return results
}

func (c *KubeObjectSizeChecker) checkObjectSize(kind, ns, name string, obj interface{}) *base.CheckResult {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil
	}

	if len(data) > WarnSizeThreshold {
		return &base.CheckResult{
			Checker:     c.Name(),
			Error:       fmt.Sprintf("%s %s/%s reaching size limit.", kind, ns, name),
			Description: fmt.Sprintf("%s %s/%s of size %d is reaching size limit. It cannot exceed 1MiB.", kind, ns, name, humanize.Bytes(uint64(len(data)))),
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
