package dns

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
	"github.com/miekg/dns"
)

type FakeDnsClient struct {
	r *dns.Msg
	e error
	m *dns.Msg
	a string
}

func (c *FakeDnsClient) Exchange(m *dns.Msg, a string) (r *dns.Msg, rtt time.Duration, err error) {
	c.m = m
	c.a = a
	return c.r, time.Duration(0), c.e
}

func TestCheckServer(t *testing.T) {
	client := &FakeDnsClient{
		r: &dns.Msg{
			MsgHdr: dns.MsgHdr{
				Rcode: dns.RcodeSuccess,
			},
		},
	}
	checker := &DnsChecker{
		client: client,
	}
	r, err := checker.checkServer(GoogleDnsServer, "www.bing.com")
	if err != nil {
		t.Errorf("expect no error but got: %+v", err)
	}
	if !r.Ok() {
		t.Errorf("expect ok but not")
	}
	if client.a != GoogleDnsServer.Server+":53" {
		t.Errorf("dns request server is wrong: %s", client.a)
	}
	if client.m.Question[0].String() != ";www.bing.com.\tIN\t A" {
		t.Errorf("wrong dns question: %s", client.m.Question[0].String())
	}
}

func TestCheckServerBadRcode(t *testing.T) {
	client := &FakeDnsClient{
		r: &dns.Msg{
			MsgHdr: dns.MsgHdr{
				Rcode: dns.RcodeServerFailure,
			},
		},
	}
	checker := &DnsChecker{
		client: client,
	}
	r, err := checker.checkServer(GoogleDnsServer, "www.bing.com")
	if err != nil {
		t.Errorf("expect no error but got: %+v", err)
	}
	if r.Ok() {
		t.Errorf("expect not ok")
	}
	if r.Error == "" || r.Description == "" ||
		!reflect.DeepEqual(r.Recommendations, GoogleDnsServer.Recommendations) ||
		!reflect.DeepEqual(r.HelpLinks, GoogleDnsServer.HelpLinks) {
		t.Errorf("unexpected result")
	}
	if client.a != GoogleDnsServer.Server+":53" {
		t.Errorf("dns request server is wrong: %s", client.a)
	}
	if client.m.Question[0].String() != ";www.bing.com.\tIN\t A" {
		t.Errorf("wrong dns question: %s", client.m.Question[0].String())
	}
}

func TestCheckServerError(t *testing.T) {
	client := &FakeDnsClient{
		e: errors.New("err"),
	}
	checker := &DnsChecker{
		client: client,
	}
	r, err := checker.checkServer(GoogleDnsServer, "www.bing.com")
	if err != nil {
		t.Errorf("expect no error but got: %+v", err)
	}
	if r.Ok() {
		t.Errorf("expect not ok")
	}
	if r.Error == "" || r.Description != "err" ||
		!reflect.DeepEqual(r.Recommendations, GoogleDnsServer.Recommendations) ||
		!reflect.DeepEqual(r.HelpLinks, GoogleDnsServer.HelpLinks) {
		t.Errorf("unexpected result")
	}
	if client.a != GoogleDnsServer.Server+":53" {
		t.Errorf("dns request server is wrong: %s", client.a)
	}
	if client.m.Question[0].String() != ";www.bing.com.\tIN\t A" {
		t.Errorf("wrong dns question: %s", client.m.Question[0].String())
	}
}

func TestGetCheckTargets(t *testing.T) {
	{
		e := &env.StaticEnvironment{
			Flags: []string{"ubuntu"},
		}
		servers := getCheckTargets(e)
		if !reflect.DeepEqual(servers, []DnsServer{GoogleDnsServer, SystemdResolvedDnsServer}) {
			t.Errorf("unexpected check targets on 'ubuntu'")
		}
	}

	{
		e := &env.StaticEnvironment{
			Flags: []string{"azure"},
		}
		servers := getCheckTargets(e)
		if !reflect.DeepEqual(servers,
			[]DnsServer{GoogleDnsServer, AzureDnsServer, AksCoreDnsServerPublic, AksCoreDnsServerInCluster}) {
			t.Errorf("unexpected check targets on 'azure'")
		}
	}

	{
		e := &env.StaticEnvironment{
			Flags: []string{""},
		}
		servers := getCheckTargets(e)
		if !reflect.DeepEqual(servers, []DnsServer{GoogleDnsServer}) {
			t.Errorf("unexpected check targets on ''")
		}
	}
}

func TestCheck(t *testing.T) {
	client := &FakeDnsClient{
		r: &dns.Msg{
			MsgHdr: dns.MsgHdr{
				Rcode: dns.RcodeSuccess,
			},
		},
	}
	checker := &DnsChecker{
		client: client,
	}

	ctx := &base.CheckContext{
		Environment: &env.StaticEnvironment{
			Flags: []string{"ubuntu"},
		},
	}
	r, err := checker.Check(ctx)
	if err != nil {
		t.Errorf("expect no error but got: %+v", err)
	}
	if len(r) != 4 {
		t.Errorf("expect 4 results but got %d", len(r))
	}
}
