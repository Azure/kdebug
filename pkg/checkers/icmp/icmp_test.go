package icmpping

import (
	"github.com/Azure/kdebug/pkg/base"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	targets := []pingTarget{{Address: "x.x.x.x"},
		{Address: "127.0.0.1"},
	}
	checker := ICMPChecker{targets: targets}
	context := &base.CheckContext{
		Environment: nil,
		KubeClient:  nil,
	}
	results, _ := checker.Check(context)
	if !isRoot() {
		if len(results) != 0 {
			t.Errorf("icmp checker unexpected results when not in root mode")
		}
	} else {
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
}
