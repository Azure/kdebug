package upgradeinspector

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
)

const checkDays = 7
const recordLimit = 50

const logPath = "/var/log/dpkg.log"

type UpgradeInspectTool struct {
	checkDays   int
	recordLimit int
}

func (t *UpgradeInspectTool) Name() string {
	return "vmrebootDetector"
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
	return result
}

func (t *UpgradeInspectTool) getAwkCmd() string {
	awkCmd := []string{
		"awk",
		"-v",
		fmt.Sprintf("tstamp=\"$(date -d \"-%v days\" +%%s)\"", t.checkDays),
		`'/install/ || /upgrade/ {dconv=gensub("-"," ","g",$1);tconv=gensub(":"," ","g",$2);dstamp=mktime(dconv" "tconv);if (dstamp >= tstamp ) { print } }'`,
		logPath,
	}
	return strings.Join(awkCmd, " ")
}
