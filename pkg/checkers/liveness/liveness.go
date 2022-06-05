package liveness

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
)

const (
	CheckerName           = "Liveness (kubelet)"
	FailedToCheckLiveness = "Failed to check liveness."
)

type LivenessChecker struct {
}

func New() *LivenessChecker {
	return &LivenessChecker{}
}

func (c *LivenessChecker) Name() string {
	return CheckerName
}

func (c *LivenessChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	results := []*base.CheckResult{}

	out, err := exec.Command("systemctl", "status", "kubelet").Output()

	if err != nil {
		results = append(results, &base.CheckResult{
			Checker:     c.Name(),
			Error:       FailedToCheckLiveness,
			Description: err.Error(),
		})
	}

	results = append(results, parseOutput(out))
	return results, nil
}

func parseOutput(output []byte) *base.CheckResult {
	rows := strings.Split(string(output), "\n")
	re := regexp.MustCompile(`active \(running\) since`)
	isActive := false
	var details string

	for _, row := range rows {
		if len(row) == 0 {
			continue
		}

		if re.MatchString(row) {
			isActive = true
			details = row
			break
		}
	}

	if isActive {
		return &base.CheckResult{
			Checker:     CheckerName,
			Description: details,
			Logs:        rows,
		}
	}

	return &base.CheckResult{
		Checker: CheckerName,
		Error:   "Kubelet is NOT running well in this node. Please check the logs for more details.",
		Logs:    rows,
	}
}
