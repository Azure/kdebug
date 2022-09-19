package diskreadonly

import (
	"errors"
	"fmt"
	"os"

	"github.com/Azure/kdebug/pkg/base"
)

const (
	HomeDirReadonlyRecommendation = "The disk enters read-only state due to underlying data integrity issues. Find out which disk your home dir is mounted on via 'df' command. Try to use 'fsck' command to fix the disk and then reboot the vm."
)

type DiskReadOnlyChecker struct {
}

func New() *DiskReadOnlyChecker {
	return &DiskReadOnlyChecker{}
}

func (c *DiskReadOnlyChecker) Name() string {
	return "DiskReadOnly"
}

func (c *DiskReadOnlyChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.New("Fail to get user home directory and check if it's read-only")
	}
	f, err := os.CreateTemp(homeDir, "testReadOnlyFile")
	var result *base.CheckResult
	if err != nil {
		result = &base.CheckResult{
			Checker:         c.Name(),
			Error:           fmt.Sprintf("Fail to create a temp file in %s", homeDir),
			Description:     err.Error(),
			Recommendations: []string{HomeDirReadonlyRecommendation},
		}
	} else {
		defer os.Remove(f.Name())
		result = &base.CheckResult{
			Checker:     c.Name(),
			Description: fmt.Sprintf("%s is not read-only", homeDir),
		}
	}
	return []*base.CheckResult{result}, nil
}
