package vmrebootdetector

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/fatih/color"
)

var helpLink = []string{
	"https://www.baeldung.com/linux/last-command",
	"https://man7.org/linux/man-pages/man1/last.1.html",
}

var explain = "This is the output of last command which is wtmp log. The columns are user, login terminal, kernel version, login time, login period\n"

const rebootCheckTimeInDay = 1

type Tool struct {
	rebootCheckTImeInDay int
}

func (t *Tool) Name() string {
	return "vmrebootDetector"
}

func New() *Tool {
	return &Tool{}
}

// Run todo: support batch mode
func (t *Tool) Run(ctx *base.ToolContext) error {
	t.parseArgument(ctx)
	return t.exec()
}

func (t *Tool) parseArgument(ctx *base.ToolContext) {
	if ctx.VmRebootDetector.CheckDays == 0 {
		t.rebootCheckTImeInDay = rebootCheckTimeInDay
	} else {
		t.rebootCheckTImeInDay = ctx.VmRebootDetector.CheckDays
	}
}

func (t *Tool) exec() error {
	sinceTime := time.Now().Add(-time.Hour * 24 * time.Duration(t.rebootCheckTImeInDay)).Format("2006-01-02 15:04:05")
	cmd := exec.Command("last", "reboot", "--since", sinceTime, "--time-format", "iso")
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(t.parseResult(string(stdout)))
	return nil
}

func (t *Tool) parseResult(result string) string {
	sb := strings.Builder{}
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
	if reboots == nil {
		sb.WriteString(color.GreenString("No reboot found in past %v days\n", t.rebootCheckTImeInDay))
	} else {
		sb.WriteString(color.RedString("Detect VM reboot\n"))
		sb.WriteString(color.YellowString(strings.Join(reboots, "\n")))
		sb.WriteString(color.GreenString("\n"))
		sb.WriteString(color.GreenString(explain))
		sb.WriteString(color.GreenString(strings.Join(helpLink, "\n")))
	}
	return sb.String()
}
