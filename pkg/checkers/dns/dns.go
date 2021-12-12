package dns

import (
	"fmt"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
	"github.com/miekg/dns"
)

const (
	PublicDnsRecommendation = "Check your public network connectivity and outbound security settings."
	CoreDnsRecommendation   = "CoreDNS pods might be down. Check their liveness using `kubectl get pods -n kube-system -o wide -l k8s-app=kube-dns`."
)

var (
	GoogleDnsServer = DnsServer{
		Name:   "Google DNS",
		Server: "8.8.8.8",
		Queries: []string{
			"www.google.com",
			"www.bing.com",
		},
		Recommendations: []string{PublicDnsRecommendation},
	}
	AzureDnsServer = DnsServer{
		Name:   "Azure DNS",
		Server: "168.63.129.16",
		Queries: []string{
			"www.google.com",
			"www.bing.com",
		},
		Recommendations: []string{
			PublicDnsRecommendation,
			"VM might be on a bad host. Try to `redeploy` it.",
		},
	}
	AksCoreDnsServerPublic = DnsServer{
		Name:   "AKS Core DNS",
		Server: "10.0.0.10",
		Queries: []string{
			"www.google.com",
			"www.bing.com",
		},
		Recommendations: []string{
			PublicDnsRecommendation,
		},
	}
	AksCoreDnsServerInCluster = DnsServer{
		Name:   "AKS Core DNS",
		Server: "10.0.0.10",
		Queries: []string{
			"kubernetes.default.svc.cluster.local",
		},
		Recommendations: []string{
			CoreDnsRecommendation,
		},
	}
	SystemdResolvedDnsServer = DnsServer{
		Name:   "systemd-resolved",
		Server: "127.0.0.53",
		Queries: []string{
			"www.google.com",
			"www.bing.com",
		},
		Recommendations: []string{
			"systemd-resolved service might not be running. Check by running `sudo systemctl status systemd-resolved`.",
		},
	}
)

type DnsServer struct {
	Name            string
	Server          string
	Queries         []string
	Recommendations []string
}

type DnsChecker struct {
	client *dns.Client
}

func New() *DnsChecker {
	return &DnsChecker{
		client: &dns.Client{
			Timeout: time.Second,
		},
	}
}

func (c *DnsChecker) Name() string {
	return "Dns"
}

func (c *DnsChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	result := []*base.CheckResult{}
	targets := getCheckTargets(ctx.Environment)
	for _, server := range targets {
		for _, query := range server.Queries {
			r, err := c.checkServer(server, query)
			if err != nil {
				return result, err
			}
			result = append(result, r)
		}
	}
	return result, nil
}

func getCheckTargets(e *env.Environment) []DnsServer {
	targets := []DnsServer{
		GoogleDnsServer,
	}

	if e.HasFlag("ubuntu") {
		targets = append(targets, SystemdResolvedDnsServer)
	}

	if e.HasFlag("azure") {
		targets = append(targets,
			AzureDnsServer,
			AksCoreDnsServerPublic,
			AksCoreDnsServerInCluster)
	}

	return targets
}

func (c *DnsChecker) checkServer(server DnsServer, query string) (*base.CheckResult, error) {
	m := new(dns.Msg)
	m.SetQuestion(query+".", dns.TypeA)
	m.RecursionDesired = true
	r, _, err := c.client.Exchange(m, server.Server+":53")
	if err != nil {
		return &base.CheckResult{
			Checker: c.Name(),
			Error: fmt.Sprintf("Fail to query domain name %s from server %s(%s)",
				query, server.Name, server.Server),
			Description:     err.Error(),
			Recommendations: server.Recommendations,
		}, nil
	}
	if r.Rcode != dns.RcodeSuccess {
		return &base.CheckResult{
			Checker: c.Name(),
			Error: fmt.Sprintf("Fail to query domain name %s from server %s(%s)", query,
				server.Name, server.Server),
			Description:     fmt.Sprintf("Unexpected rcode: %d", r.Rcode),
			Recommendations: server.Recommendations,
		}, nil
	}
	return &base.CheckResult{
		Checker: c.Name(),
		Description: fmt.Sprintf("Successfully query domain name %s from server %s(%s)",
			query, server.Name, server.Server),
	}, nil
}
