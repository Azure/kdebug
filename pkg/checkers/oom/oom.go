package oom

import (
	"bufio"
	"fmt"
	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strings"
)

const (
	logPath   = "/var/log/kern.log"
	oomKeyStr = "Memory cgroup out of memory"
)

var helpLink = []string{
	"https://www.kernel.org/doc/gorman/html/understand/understand016.html",
	"https://stackoverflow.com/questions/18845857/what-does-anon-rss-and-total-vm-mean",
	"https://medium.com/tailwinds-navigator/kubernetes-tip-how-does-oomkilled-work-ba71b135993b",
}

var oomRegex = regexp.MustCompile("^(.*:.{2}:.{2}) .* process (.*) \\((.*)\\) .* anon-rss:(.*), file-rss.* oom_score_adj:(.*)")

type OOMChecker struct {
	kernLogPath string
}

func (c *OOMChecker) Name() string {
	return "OOM"
}

func New() *OOMChecker {
	//todo: support other logpath
	return &OOMChecker{
		kernLogPath: logPath,
	}
}

func (c *OOMChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	var results []*base.CheckResult
	oomResult, err := c.checkOOM(ctx)
	if err != nil {
		log.Warnf("error while checking oom:%v\n", err)
	}
	results = append(results, oomResult)
	return results, nil
}

func (c *OOMChecker) checkOOM(ctx *base.CheckContext) (*base.CheckResult, error) {
	result := &base.CheckResult{
		Checker: c.Name(),
	}
	if !envCheck(ctx.Environment) {
		result.Description = fmt.Sprint("Skip oom check in non-linux os")
		return result, nil
	}
	oomInfos, err := c.getAndParseOOMLog()
	if err != nil {
		return nil, err
	} else if len(oomInfos) > 0 {
		result.Error = strings.Join(oomInfos, "\n")
		result.Description = "Detect process oom killed"
		result.HelpLinks = helpLink
	} else {
		result.Description = "No OOM found in recent kernlog."
	}
	return result, nil
}
func (c *OOMChecker) getAndParseOOMLog() ([]string, error) {
	file, err := os.Open(c.kernLogPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var oomInfos []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tmp := scanner.Text()
		//todo: more sophisticated OOM context
		if strings.Contains(tmp, oomKeyStr) {
			oomInfo, err := parseOOMContent(tmp)
			if err != nil {
				return nil, err
			} else {
				oomInfos = append(oomInfos, oomInfo)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return oomInfos, nil
}

func parseOOMContent(content string) (string, error) {
	match := oomRegex.FindStringSubmatch(content)
	if len(match) != 6 {
		err := fmt.Errorf("Can't parse oom content:%s \n", content)
		return "", err
	} else {
		return fmt.Sprintf("progress:[%s %s] is OOM kill at time [%s]. [rss:%s] [oom_score_adj:%s]\n", match[2], match[3], match[1], match[4], match[5]), nil
	}
}

func envCheck(environment env.Environment) bool {
	//todo:support other os
	return environment.HasFlag("ubuntu")
}
