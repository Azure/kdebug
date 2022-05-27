package vmreboot

import (
	"fmt"
	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
	"testing"
)

func TestReboot(t *testing.T) {
	environment := &env.StaticEnvironment{
		Flags: []string{"ubuntu"},
	}
	if !envCheck(env.GetEnvironment()) {
		fmt.Println("skip vm reboot test")
		return
	}
	checker := VMRebootChecker{}
	_, err := checker.Check(&base.CheckContext{
		Environment: environment,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestRebootParser(t *testing.T) {
	lastContent := "reboot   system boot  5.4.0-1074-azure 2022-05-27T04:51:43+0000   still running\nreboot   system boot  5.4.0-1074-azure 2022-04-04T07:49:09+0000 - 2022-04-20T17:12:20+0000 (16+09:23)\n\nwtmp begins 2022-04-04T07:47:27+0000\n"
	checker := VMRebootChecker{}
	checkResult := checker.parseResult(lastContent)
	if checkResult.Error == "" {
		t.Error("VMRebootCheck failed to parse reboot result")
	}
}
