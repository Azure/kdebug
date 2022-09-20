package vmrebootdetector

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/kdebug/pkg/base"
)

func TestReboot(t *testing.T) {
	tool := Tool{}
	err := tool.Run(&base.ToolContext{
		Config: &Config{},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestRebootParser(t *testing.T) {
	lastContent := "reboot   system boot  5.4.0-1074-azure 2022-05-27T04:51:43+0000   still running\nreboot   system boot  5.4.0-1074-azure 2022-04-04T07:49:09+0000 - 2022-04-20T17:12:20+0000 (16+09:23)\n\nwtmp begins 2022-04-04T07:47:27+0000\n"
	tool := Tool{}
	result := tool.parseResult(lastContent)
	fmt.Println(result)
	if !strings.Contains(result, "Detect") {
		t.Error("VMRebootCheck failed to parse reboot result")
	}
}
