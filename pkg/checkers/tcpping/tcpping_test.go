package tcpping

import (
	"fmt"
	"github.com/Azure/kdebug/pkg/base"
	"math/rand"
	"net"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	checker := &TCPChecker{
		dialer: net.Dialer{
			Timeout: timeOut,
		},
		targets: []pingEndpoint{
			{
				ServerAddress: "fooTest",
				Name:          fmt.Sprintf("%d.kdebug:80", rand.Int()),
			},
		},
	}
	context := &base.CheckContext{}
	results, err := checker.Check(context)
	if err != nil {
		t.Errorf("check fail %v\n", err)
	}
	for _, result := range results {
		if strings.Contains(result.Description, "fooTest") {
			if result.Error == "" {
				t.Errorf("fooTest didn't fail")
			}
		}
		if strings.Contains(result.Description, "google") {
			if result.Error != "" {
				t.Errorf("google test fail %v\n", result.Error)
			}
		}
	}
}
