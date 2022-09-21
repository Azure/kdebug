package upgradeinspector

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
	"github.com/fatih/color"
)

const logPath = "/var/log/dpkg.log"

const suggestion = "You can check '/var/log/dpkg.log' and '/var/log/apt/history.log' for further detail."

var columns = []string{
	"Timestamp",
	"Package",
	"OldVer",
	"NewVer",
}

type UpgradeInspectTool struct {
	checkDays   int
	recordLimit int
}

func (t *UpgradeInspectTool) Name() string {
	return "upgradeinspector"
}

func New() *UpgradeInspectTool {
	return &UpgradeInspectTool{}
}

func (t *UpgradeInspectTool) Run(ctx *base.ToolContext) error {
	t.parseArgument(ctx)
	if !envCheck(ctx.Environment) {
		fmt.Println(color.YellowString("Skip upgrade inspect in non ubuntu/debian os"))
		return nil
	}
	return t.exec()
}

func (t *UpgradeInspectTool) parseArgument(ctx *base.ToolContext) {
	t.checkDays = ctx.UpgradeInspector.CheckDays
	t.recordLimit = ctx.UpgradeInspector.RecordLimit
}

func (t *UpgradeInspectTool) exec() error {
	cmd := exec.Command("grep", " upgrade ", logPath)
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(t.parseResult(string(stdout)))
	fmt.Println(color.YellowString("\n%v\n", suggestion))
	return nil
}

func (t *UpgradeInspectTool) parseResult(result string) string {
	sb := strings.Builder{}
	logs := strings.Split(result, "\n")
	logNum := len(logs) - 1

	if logNum == 0 {
		sb.WriteString(color.GreenString("\nNo package upgrade log found\n"))
	} else {
		sb.WriteString(fmt.Sprintf("\n%-19s\t%-30s\t%-20s\t%-20s\n\n", columns[0], columns[1], columns[2], columns[3]))
	}

	cutTime := time.Now().AddDate(0, 0, -t.checkDays)
	for i := 0; i < logNum && i < t.recordLimit; i++ {
		strs := strings.Split(logs[i], " ")
		logTime, err := time.Parse("2006-01-02 15:04:05", fmt.Sprintf(`%s %s`, strs[0], strs[1]))
		if err == nil && logTime.After(cutTime) {
			sb.WriteString(fmt.Sprintf("%v-%v\t%-30s\t%-20s\t%-20s\n", strs[0], strs[1], strs[3], strs[4], strs[5]))
		}
	}
	if t.recordLimit < logNum {
		sb.WriteString(color.YellowString("\n%v package(s) omitted\n", logNum-t.recordLimit))
	}
	return sb.String()
}

func envCheck(environment env.Environment) bool {
	return environment.HasFlag("ubuntu") || environment.HasFlag("debian")
}
