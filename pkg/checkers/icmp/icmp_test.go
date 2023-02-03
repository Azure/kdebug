package icmpping

import (
	"os"
	"strings"
	"testing"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
)

func TestICMPCheckRoot(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Must run with root")
		return
	}
	targets := []pingTarget{{Address: "x.x.x.x"},
		{Address: "127.0.0.1"},
	}
	checker := ICMPChecker{targets: targets}
	context := &base.CheckContext{
		Environment: &env.StaticEnvironment{
			Flags: []string{"root"},
		},
		KubeClient: nil,
	}
	results, _ := checker.Check(context)
	for _, result := range results {
		if strings.Contains(result.Description, "x.x.x.x") {
			if result.Error == "" {
				t.Errorf("ping x.x.x.x should fail")
			}
		}
		if strings.Contains(result.Description, "127.0.0.1") {
			if result.Error != "" {
				t.Errorf("ping 127.0.0.1 failed %v\n", result.Error)
			}
		}
	}
}

func TestICMPCheckNonRoot(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("Must run with non-root")
		return
	}

	targets := []pingTarget{{Address: "x.x.x.x"},
		{Address: "127.0.0.1"},
	}
	checker := ICMPChecker{targets: targets}
	context := &base.CheckContext{
		Environment: &env.StaticEnvironment{
			Flags: []string{},
		},
		KubeClient: nil,
	}
	results, _ := checker.Check(context)
	if len(results) != 0 {
		t.Errorf("icmp checker unexpected results when not in root mode")
	}
}
