package vmreboot

import (
	"bufio"
	"fmt"
	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
	"os/exec"
	"strings"
	"time"
)

var helpLink = []string{
	"https://www.baeldung.com/linux/last-command",
	"https://man7.org/linux/man-pages/man1/last.1.html",
}

const rebootCheckTimeInDay = 1

type VMRebootChecker struct {
}

func (c *VMRebootChecker) Name() string {
	return "VMReboot"
}

func New() *VMRebootChecker {
	return &VMRebootChecker{}
}

func (c *VMRebootChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	var results []*base.CheckResult
	rebootResult, err := c.checkReboot(ctx)
	if err != nil {
		return nil, err
	}
	results = append(results, rebootResult)
	return results, nil
}

func (c *VMRebootChecker) checkReboot(ctx *base.CheckContext) (*base.CheckResult, error) {
	result := &base.CheckResult{
		Checker: c.Name(),
	}
	if !envCheck(ctx.Environment) {
		result.Description = fmt.Sprint("Skip reboot check in non-linux os")
		return result, nil
	}
	sinceTime := time.Now().Add(-time.Hour * 24 * rebootCheckTimeInDay)
	lastArg := fmt.Sprintf("reboot --since %s --time-format iso", sinceTime.Format("2006-01-02 15:04:05"))
	cmd := exec.Command("last", lastArg)

	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return c.parseResult(string(stdout)), nil
}

func (c *VMRebootChecker) parseResult(result string) *base.CheckResult {
	scanner := bufio.NewScanner(strings.NewReader(result))
	var reboots []string
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			break
		} else {
			reboots = append(reboots, text)
		}
	}
	checkResult := &base.CheckResult{
		Checker: c.Name(),
	}
	if reboots == nil {
		checkResult.Description = fmt.Sprintf("No reboot found in past %v days", rebootCheckTimeInDay)
	} else {
		checkResult.Description = "Detect VM reboot"
		checkResult.Error = strings.Join(reboots, "\n")
		checkResult.HelpLinks = helpLink
	}
	return checkResult
}

func envCheck(environment env.Environment) bool {
	//todo:support other os
	return environment.HasFlag("ubuntu")
}
