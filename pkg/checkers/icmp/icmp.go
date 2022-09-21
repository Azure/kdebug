package icmpping

import (
	"context"
	"errors"
	"fmt"
	"os/user"
	"time"

	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/kdebug/pkg/base"
)

var PublicTargets = []pingTarget{{
	Address:  "8.8.8.8",
	Name:     "GoogleDns",
	Category: "Public",
}}

type ICMPChecker struct {
	targets []pingTarget
}

type pingTarget struct {
	Address  string
	Name     string
	Category string
}

func New() *ICMPChecker {
	return &ICMPChecker{}
}

func (c *ICMPChecker) Name() string {
	return "icmp"
}

func (c *ICMPChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	var results []*base.CheckResult
	if !isRoot() {
		log.Warn("Not root. Skip icmp checker")
		return results, nil
	}
	if !ctx.Environment.HasFlag("azure") {
		c.targets = append(c.targets, PublicTargets...)
	}
	c.targets = append(c.targets, getInClusterTargets(ctx)...)
	resultChan := make(chan *base.CheckResult, len(c.targets))
	for _, target := range c.targets {
		go func(pingTarget pingTarget) {
			result := &base.CheckResult{
				Checker: c.Name(),
			}
			err := pingOne(pingTarget.Address)
			if err != nil {
				result.Error = err.Error()
				result.Description = fmt.Sprintf("ping %s %s[%s] failed", pingTarget.Category, pingTarget.Address, pingTarget.Name)
			} else {
				result.Description = fmt.Sprintf("ping %s %s[%s] succeeded", pingTarget.Category, pingTarget.Address, pingTarget.Name)
			}
			resultChan <- result

		}(target)
	}
	for i := 0; i < len(c.targets); i++ {
		result := <-resultChan
		results = append(results, result)
	}
	return results, nil
}

func pingOne(ip string) error {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return err
	}

	pinger.Count = 3
	pinger.Interval = time.Millisecond * 20
	pinger.Timeout = time.Millisecond * 1000
	err = pinger.Run()
	if err != nil {
		return err
	}
	stats := pinger.Statistics()
	if stats.PacketsRecv <= 0 {
		return errors.New("ping receive no reply")
	}
	return nil
}

func getInClusterTargets(ctx *base.CheckContext) []pingTarget {
	var targets []pingTarget
	if ctx.KubeClient != nil {
		nodes, err := ctx.KubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Warnf("get nodes error %v. Skip cluster nodes", err)
		} else {
			for _, node := range nodes.Items {
				address := getNodeAddress(node)
				if address != "" {
					targets = append(targets, pingTarget{
						Address:  address,
						Category: "ClusterNode",
						Name:     node.Name,
					})
				}
			}
		}
		pods, err := ctx.KubeClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Warnf("get nodes error %v. Skip pod", err)
		} else {
			for _, pod := range pods.Items {
				address := pod.Status.PodIP
				if address != "" {
					targets = append(targets, pingTarget{
						Address:  address,
						Category: "Pod",
						Name:     pod.Name,
					})
				}
			}
		}
	}

	return targets
}

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Warnf("Get user error %v.", err)
		return false
	}
	return currentUser.Username == "root"
}

func getNodeAddress(node v1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeHostName {
			return address.Address
		}
	}
	return ""
}
