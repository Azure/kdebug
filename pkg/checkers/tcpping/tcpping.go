package main

import (
	"fmt"
	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
	"net"
	"time"
)

const timeOut = 2 * time.Second

const (
	GoogleTarget    = "google.com:443"
	AzureIMDSTarget = "169.254.169.254:80"
)

func (t *TCPChecker) ping(serverAddr string) error {
	conn, err := t.dialer.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func main() {
	tcp := New()
	tcp.ping(GoogleTarget)
	tcp.ping(AzureIMDSTarget)
}

type TCPChecker struct {
	dialer net.Dialer
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
	targets := getCheckTargets(ctx.Environment)
	var result *base.CheckResult
	for _, serverAddress := range targets {
		err := t.ping(serverAddress)
		if err != nil {
			result = &base.CheckResult{
				Checker: t.Name(),
				Error:   err.Error(),
				Description: fmt.Sprintf("Fail to establish tcp connection to  %s.",
					serverAddress),
			}
		} else {
			result = &base.CheckResult{
				Checker:     t.Name(),
				Description: fmt.Sprintf("successfully establish tco connection to %s", serverAddress),
			}
		}
		results = append(results, result)
	}

	return results, nil
}

func getCheckTargets(e env.Environment) []string {
	targets := []string{GoogleTarget}
	if e.HasFlag("azure") {
		targets = append(targets, AzureIMDSTarget)
	}
	return targets
}
