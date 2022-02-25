package oom

import (
	"bufio"
	"fmt"
	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
	"os"
	"regexp"
	"strings"
)

const logPath = "/var/log/kern.log"
const oomKeyStr = "Memory cgroup out of memory"

//const testString = "Feb 22 16:15:02 k8s-ingress-11186066-z1-vmss0000B3 kernel: [989751.247878] Memory cgroup out of memory: Killed process 3841 (nginx) total-vm:240652kB, anon-rss:130344kB, file-rss:5212kB, shmem-rss:208kB, UID:101 pgtables:332kB oom_score_adj:986\n"

var oomRegex = regexp.MustCompile("^(.*:.{2}:.{2}) .* process (.*) \\((.*)\\) .* anon-rss:(.*), file-rss")

type OOMChecker struct {
	kernLogPath string
}

func (c *OOMChecker) Name() string {
	return "OOM"
}

func New() *OOMChecker {
	return &OOMChecker{
		kernLogPath: logPath,
	}
}

func (c *OOMChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	var results []*base.CheckResult
	oomResult, _ := c.checkOOM(ctx)
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
		result.Description = fmt.Sprintf("Fail to check OOM because of unexpected error:%v", err)
	} else if len(oomInfos) > 0 {
		result.Error = strings.Join(oomInfos, "\n")
		result.Description = "Detect process oom killed"
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
	if len(match) != 5 {
		err := fmt.Errorf("Can't parse oom content:%s \n", content)
		return "", err
	} else {
		return fmt.Sprintf("progress:[%s %s] is OOM kill at time [%s]. [rss:%s]\n", match[2], match[3], match[1], match[4]), nil
	}
}

func envCheck(environment env.Environment) bool {
	//should include more flags
	return environment.HasFlag("ubuntu")
}
