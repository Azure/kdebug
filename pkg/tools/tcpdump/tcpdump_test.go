package tcpdump

import (
	"testing"
)

func TestParseIPAndPort_Success(t *testing.T) {
	ip, port := ParseIPAndPort("192.168.1.1:7")
	if ip != "192.168.1.1" || port != "7" {
		t.Errorf("Parsing 192.168.1.1:7 and expect ip is 192.168.1.1 and port is 7 but got %s and %s", ip, port)
	}

	ip, port = ParseIPAndPort("192.168.1.1")
	if ip != "192.168.1.1" || len(port) != 0 {
		t.Errorf("Parsing 192.168.1.1 and expect ip is 192.168.1.1 and no port but got %s and %s", ip, port)
	}

	ip, port = ParseIPAndPort("192.168.1.1:")
	if ip != "192.168.1.1" || len(port) != 0 {
		t.Errorf("Parsing 192.168.1.1: and expect ip is 192.168.1.1 and no port but got %s and %s", ip, port)
	}

	ip, port = ParseIPAndPort(":80")
	if len(ip) != 0 || port != "80" {
		t.Errorf("Parsing :80 and expect no ip and port is 80 but got %s and %s", ip, port)
	}
}

func TestGenerateTcpdumpParamerters_Success(t *testing.T) {
	tcpdumptool := New()

	config := &Config{"192.168.1.1:1", "23.32.10.2:80", ":443", "19920", true}
	tcpdumptool.ParseParameters(config)
	parameter := tcpdumptool.GenerateTcpdumpParamerters()

	expected := "-nvvv src 192.168.1.1 and src port 1 and dst 23.32.10.2 and dst port 80 and port 443 and tcp"
	if parameter != expected {
		t.Errorf("Generate parameter is expected to be %s but actually %s", expected, parameter)
	}
}
