package icmpping

import (
	"errors"
	"fmt"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
)

var PublicTargets = []pingTarget{
	{
		Address:        "8.8.8.8",
		Name:           "GoogleDns",
		Recomendations: []string{"Google DNS is not reachable. Check firewall settings if this is not desired."},
	},
	{
		Address:        "10.0.0.10",
		Name:           "ClusterDns",
		Recomendations: []string{"Cluster CoreDNS is not reachable. Check CoreDNS pods and network settings."},
	},
}

type ICMPChecker struct {
	targets []pingTarget
}

type pingTarget struct {
	Address        string
	Name           string
	Recomendations []string
}

func New() *ICMPChecker {
	return &ICMPChecker{}
}

func (c *ICMPChecker) Name() string {
	return "icmp"
}

func (c *ICMPChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	var results []*base.CheckResult
	// TODO: Invoke `ping` command if non-root
	if !ctx.Environment.HasFlag("root") {
		log.Debug("Not root. Skip ICMP checker")
		return results, nil
	}
	if !ctx.Environment.HasFlag("azure") {
		c.targets = append(c.targets, PublicTargets...)
	}
	if ctx.KubeClient != nil {

	}
	resultChan := make(chan *base.CheckResult, len(c.targets))
	for _, target := range c.targets {
		go func(pingTarget pingTarget) {
			result := &base.CheckResult{
				Checker: c.Name(),
			}
			err := pingOne(pingTarget.Address)
			if err != nil {
				result.Error = err.Error()
				result.Description = fmt.Sprintf("ping %s[%s] failed", pingTarget.Address, pingTarget.Name)
				result.Recommendations = pingTarget.Recomendations
			} else {
				result.Description = fmt.Sprintf("ping %s[%s] succeeded", pingTarget.Address, pingTarget.Name)
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
		return errors.New("ping receives no reply")
	}
	return nil
}
