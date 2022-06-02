package liveness

import (
	"github.com/Azure/kdebug/pkg/base"
)

type LivenessChecker struct {
}

func New() *LivenessChecker {
	return &LivenessChecker{}
}

func (c *LivenessChecker) Name() string {
	return "Liveness"
}

func (c *LivenessChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	result := []*base.CheckResult{}
	return result, nil
}
