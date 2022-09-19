package tcpping

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Azure/kdebug/pkg/base"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const timeOut = 2 * time.Second

const (
	GoogleTarget = "google.com:443"
)

type pingEndpoint struct {
	ServerAddress string
	Name          string
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
				Description: color.RedString(fmt.Sprintf("Fail to establish tcp connection to %s (%s).",
					pingTarget.ServerAddress, pingTarget.Name)),
			}
		} else {
			result = &base.CheckResult{
				Checker:     t.Name(),
				Description: color.GreenString(fmt.Sprintf("successfully establish tcp connection to %s (%s)", pingTarget.ServerAddress, pingTarget.Name)),
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
		services, err := getExternalServicePingEndpoint(c)
		if err != nil {
			log.Warn(fmt.Sprintf("fetch external endpoint error %v.Skip those checks", err))
		} else {
			targets = append(targets, services...)
		}
	}
	return targets
}

func getExternalServicePingEndpoint(c *base.CheckContext) ([]pingEndpoint, error) {
	services, err := c.KubeClient.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var lbServices []pingEndpoint
	for _, service := range services.Items {
		for _, port := range service.Spec.Ports {
			if port.Protocol == v1.ProtocolTCP {
				address := ""
				if service.Spec.Type == "LoadBalancer" {
					address = service.Spec.LoadBalancerIP
				} else if service.Spec.Type == "ClusterIP" {
					//address = service.Spec.ClusterIP
				}
				if address != "" {
					serverUrl := fmt.Sprintf("%s:%d", address, port.Port)
					lbServices = append(lbServices, pingEndpoint{
						ServerAddress: serverUrl,
						Name:          service.Name,
					})
				}
			}
		}

	}
	return lbServices, nil
}
