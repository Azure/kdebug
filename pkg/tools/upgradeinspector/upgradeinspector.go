package upgradeinspector

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/fatih/color"
)

const checkDays = 7
const recordLimit = 50

const logPath = "/var/log/dpkg.log"

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

// Run todo: support batch mode
func (t *UpgradeInspectTool) Run(ctx *base.ToolContext) error {
	t.parseArgument(ctx)
	return t.exec()
}

func (t *UpgradeInspectTool) parseArgument(ctx *base.ToolContext) {
	t.checkDays = checkDays
	t.recordLimit = recordLimit

	if ctx.UpgradeInspector.CheckDays != 0 {
		t.checkDays = ctx.UpgradeInspector.CheckDays
	}
	if ctx.UpgradeInspector.RecordLimit != 0 {
		t.recordLimit = ctx.UpgradeInspector.RecordLimit
	}
}

func (t *UpgradeInspectTool) exec() error {
	cmd := exec.Command("bash", "-c", t.getAwkCmd())
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(t.parseResult(string(stdout)))
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

	for i := 0; i < logNum && i < t.recordLimit; i++ {
		strs := strings.Split(logs[i], " ")
		sb.WriteString(fmt.Sprintf("%v-%v\t%-30s\t%-20s\t%-20s\n", strs[0], strs[1], strs[3], strs[4], strs[5]))
	}
	if t.recordLimit < logNum {
		sb.WriteString(color.YellowString("\n%v package(s) omitted\n", logNum-t.recordLimit))
	}
	return sb.String()
}

func (t *UpgradeInspectTool) getAwkCmd() string {
	awkCmd := []string{
		"awk",
		"-v",
		fmt.Sprintf("tstamp=\"$(date -d \"-%v days\" +%%s)\"", t.checkDays),
		`'/upgrade/ {dconv=gensub("-"," ","g",$1);tconv=gensub(":"," ","g",$2);dstamp=mktime(dconv" "tconv);if (dstamp >= tstamp) { print } }'`,
		logPath,
	}
	return strings.Join(awkCmd, " ")
}
