package tcpping

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Azure/kdebug/pkg/base"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const KubernetesServiceHost = "KUBERNETES_SERVICE_HOST"
const timeOut = 1000 * time.Millisecond

const (
	GoogleTarget = "google.com:443"
)

type pingEndpoint struct {
	ServerAddress string
	Name          string
	NameSpace     string
}

func (t *TCPChecker) ping(serverAddr string) error {
	conn, err := t.dialer.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

type TCPChecker struct {
	dialer  net.Dialer
	targets []pingEndpoint
}

func New() *TCPChecker {
	return &TCPChecker{
		dialer: net.Dialer{
			Timeout: timeOut,
		},
	}
}

func (t *TCPChecker) Name() string {
	return "TcpChecker"
}

func (t *TCPChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	var results []*base.CheckResult
	targets := append(t.targets, getCheckTargets(ctx)...)
	var result *base.CheckResult
	for _, pingTarget := range targets {
		err := t.ping(pingTarget.ServerAddress)
		if err != nil {
			result = &base.CheckResult{
				Checker: t.Name(),
				Error:   err.Error(),
				Description: color.RedString(fmt.Sprintf("Fail to establish tcp connection to %s (%s) in namespace %s.",
					pingTarget.ServerAddress, pingTarget.Name, pingTarget.NameSpace)),
				Recommendations: []string{"This might be expected for example the firewall blocks the traffic"},
			}
		} else {
			result = &base.CheckResult{
				Checker:     t.Name(),
				Description: color.GreenString(fmt.Sprintf("Successfully establish tcp connection to %s (%s) in namespace %s", pingTarget.ServerAddress, pingTarget.Name, pingTarget.NameSpace)),
			}
		}
		results = append(results, result)
	}

	return results, nil
}

func getCheckTargets(c *base.CheckContext) []pingEndpoint {
	var targets []pingEndpoint
	targets = append(targets, pingEndpoint{Name: "Google", ServerAddress: GoogleTarget})

	if c.KubeClient != nil {
		services, err := getServicePingEndpoint(c)
		if err != nil {
			log.Warn(fmt.Sprintf("Fetch cluster servuce ping endpoint error %v.Skip those checks", err))
		} else {
			targets = append(targets, services...)
		}
	}
	return targets
}

func getServicePingEndpoint(c *base.CheckContext) ([]pingEndpoint, error) {
	services, err := c.KubeClient.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	isInKubernetes := checkIfInsideKubernetes()
	if err != nil {
		return nil, err
	}
	var pingEndpoints []pingEndpoint
	for _, service := range services.Items {
		for _, port := range service.Spec.Ports {
			if port.Protocol == v1.ProtocolTCP {
				address := formatIP(service.Spec.LoadBalancerIP)
				if address == "" && len(service.Status.LoadBalancer.Ingress) > 0 {
					address = formatIP(service.Status.LoadBalancer.Ingress[0].IP)
				}
				if address == "" && isInKubernetes {
					address = formatIP(service.Spec.ClusterIP)
				}
				if address != "" {
					serverUrl := fmt.Sprintf("%s:%d", address, port.Port)
					pingEndpoints = append(pingEndpoints, pingEndpoint{
						ServerAddress: serverUrl,
						Name:          service.Name,
						NameSpace:     service.Namespace,
					})
				}
			}
		}

	}
	return pingEndpoints, nil
}

func formatIP(address string) string {
	if address == "" || address == "None" {
		return ""
	}
	if strings.Contains(address, ":") {
		return fmt.Sprintf("[%s]", address)
	} else {
		return address
	}
}

func checkIfInsideKubernetes() bool {
	//check if in a pod
	for _, e := range os.Environ() {
		if strings.Contains(e, KubernetesServiceHost) {
			return true
		}
	}
	// check in a host vm
	processes, err := process.Processes()
	if err != nil {
		log.Warn(fmt.Sprintf("List process error %v. Skip in-cluster tcp checking\n", err))
		return false
	}
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			log.Warn(fmt.Sprintf("List process error %v. Skip in-cluster tcp checking\n", err))
			return false
		}
		if name == "kubelet" {
			return true
		}
	}
	return false
}
