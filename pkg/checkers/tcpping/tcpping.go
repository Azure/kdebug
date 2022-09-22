package tcpping

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/kdebug/pkg/base"
)

const KubernetesServiceHost = "KUBERNETES_SERVICE_HOST"
const TimeOut = 1000 * time.Millisecond

var PublicTargets = []pingEndpoint{
	{
		ServerAddress: "www.google.com:443",
		Name:          "Google",
		NameSpace:     "",
	},
}

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
	defer conn.Close()
	conn.(*net.TCPConn).SetLinger(0)
	return nil
}

type TCPChecker struct {
	dialer  net.Dialer
	targets []pingEndpoint
}

func New() *TCPChecker {
	return &TCPChecker{
		dialer: net.Dialer{
			Timeout: TimeOut,
		},
	}
}

func (t *TCPChecker) Name() string {
	return "TcpChecker"
}

func (t *TCPChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	var results []*base.CheckResult
	targets := append(t.targets, getCheckTargets(ctx)...)
	resultChan := make(chan *base.CheckResult, len(targets))
	for _, pingTarget := range targets {
		go func(target pingEndpoint) {
			result := &base.CheckResult{
				Checker: t.Name(),
			}
			err := t.ping(target.ServerAddress)
			sb := strings.Builder{}
			if err != nil {
				sb.WriteString(fmt.Sprintf("Fail to establish tcp connection to %s (%s) ",
					target.ServerAddress, target.Name))
				result.Error = err.Error()
				result.Recommendations = []string{"Check firewall settings if this is not expected."}
			} else {
				sb.WriteString(fmt.Sprintf("Successfully establish tcp connection to %s (%s)", target.ServerAddress, target.Name))
			}
			if target.NameSpace != "" {
				sb.WriteString(fmt.Sprintf(" in namespace %s", target.NameSpace))
			}
			sb.WriteString("\n")
			result.Description = sb.String()
			resultChan <- result
		}(pingTarget)
	}
	for i := 0; i < len(targets); i++ {
		result := <-resultChan
		results = append(results, result)
	}
	return results, nil
}

func getCheckTargets(c *base.CheckContext) []pingEndpoint {
	var targets []pingEndpoint
	targets = append(targets, PublicTargets...)
	if c.KubeClient != nil {
		services, err := getServicePingEndpoint(c)
		if err != nil {
			log.Warnf("Fetch cluster service ping endpoint error %v.Skip those checks", err)
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
		log.Warnf("List process error %v. Skip in-cluster tcp checking\n", err)
		return false
	}
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			log.Warnf("List process error %v. Skip in-cluster tcp checking\n", err)
			return false
		}
		if name == "kubelet" {
			return true
		}
	}
	return false
}
