package upgradeinspector

import (
	"fmt"
	"testing"
	"time"

	"github.com/Azure/kdebug/pkg/base"
)

func TestUpgradeParser_Success(t *testing.T) {
	upgradeInspectTool := New()

	ctx := &base.ToolContext{
		Config: &Config{CheckDays: 7, RecordLimit: 10},
	}

	upgradeInspectTool.parseArgument(ctx)

	logTime := time.Now().AddDate(0, 0, -1)
	dateStr := logTime.Format("2006-01-02")

	logs := fmt.Sprintf("%s 17:12:13 upgrade libubsan1:amd64 12-20220319-1ubuntu1 12.1.0-2ubuntu1~22.04\n", dateStr) +
		fmt.Sprintf("%s 17:12:13 upgrade gcc-12-base:amd64 12-20220319-1ubuntu1 12.1.0-2ubuntu1~22.04\n", dateStr)

	expected := fmt.Sprintf("\n%-19s\t%-40s\t%-30s\t%-30s\n\n", "Timestamp", "Package", "OldVer", "NewVer") +
		fmt.Sprintf("%v-%v\t%-40s\t%-30s\t%-30s\n", dateStr, "17:12:13", "libubsan1:amd64", "12-20220319-1ubuntu1", "12.1.0-2ubuntu1~22.04") +
		fmt.Sprintf("%v-%v\t%-40s\t%-30s\t%-30s\n", dateStr, "17:12:13", "gcc-12-base:amd64", "12-20220319-1ubuntu1", "12.1.0-2ubuntu1~22.04")

	output := upgradeInspectTool.parseResult(logs)

	if output != expected {
		t.Errorf("UpgradeInspectTool parser output is expected to be\n%s\n, but got\n%s\n", expected, output)
	}
}
