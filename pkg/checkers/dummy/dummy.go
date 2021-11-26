package dummy

import (
	"os"

	"github.com/Azure/kdebug/pkg/base"
)

type DummyChecker struct {
}

var okResult = base.CheckResult{
	Checker: "Dummy",
}

var failResult = base.CheckResult{
	Checker:     "Dummy",
	Error:       "Dummy failure",
	Description: "This is a dummy failure",
	Recommandations: []string{
		"Remove environment variable `KDEBUG_DUMMY_FAIL`.",
	},
}

func (c *DummyChecker) Name() string {
	return "Dummy"
}

func (c *DummyChecker) Check(_ *base.CheckContext) ([]*base.CheckResult, error) {
	if os.Getenv("KDEBUG_DUMMY_FAIL") == "1" {
		return []*base.CheckResult{&failResult}, nil
	} else {
		return []*base.CheckResult{&okResult}, nil
	}
}
