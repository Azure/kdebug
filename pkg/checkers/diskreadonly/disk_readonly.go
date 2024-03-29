package diskreadonly

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/kdebug/pkg/base"
)

const (
	Reason                = "The filesystem mignt enter read-only state due to underlying data integrity issues."
	GeneralRecommendation = "Find out which filesystem your home dir is mounted on via 'df' command. Try to use 'fsck' command to fix the filesystem and then reboot the vm."
)

var helpLink = []string{
	"linux.die.net/man/8/mount",
	"linux.die.net/man/8/fsck",
	"https://askubuntu.com/a/197468",
}

type DiskReadOnlyChecker struct {
}

func New() *DiskReadOnlyChecker {
	return &DiskReadOnlyChecker{}
}

func (c *DiskReadOnlyChecker) Name() string {
	return "DiskReadOnly"
}

func (c *DiskReadOnlyChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	if !ctx.Environment.HasFlag("linux") {
		// This checker is only valid on Linux.
		log.Debugf("Skip %s checker in non-linux os", c.Name())
		return []*base.CheckResult{}, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("Fail to get user home dir. %w", err)
	}

	f, err := os.CreateTemp(homeDir, "testReadOnlyFile")
	var result *base.CheckResult
	if err != nil {
		var recommendation string
		if strings.Contains(strings.ToLower(err.Error()), "read-only") {
			mountSrc, mountTarget, findMntErr := getMountSrcAndTarget(homeDir)
			if findMntErr != nil {
				log.Warnf("Fail to find mount src for %s: %s", homeDir, findMntErr)
				recommendation = fmt.Sprintf("%s%s", Reason, GeneralRecommendation)
			} else {
				recommendation = fmt.Sprintf("%s Try to use 'fsck' command to fix the %s mounted on %s and then reboot the vm.", Reason, mountSrc, mountTarget)
			}
			result = &base.CheckResult{
				Checker:         c.Name(),
				Error:           "Disk might be read-only",
				Description:     fmt.Sprintf("Cannot create a temp file in %s due to %s", homeDir, err),
				Recommendations: []string{recommendation},
				HelpLinks:       []string{},
			}
		} else {
			return nil, fmt.Errorf("Fail to create a temp file in %s due to unexpected error: %w", homeDir, err)
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

func getMountSrcAndTarget(homeDir string) (string, string, error) {
	findMntCmd := exec.Command("findmnt", "--target", homeDir, "--output", "SOURCE,TARGET", "--noheadings")
	mountDescription, err := findMntCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("Fail to find the filesystem of %s with command '%s': %w",
			homeDir, findMntCmd.String(), err)
	} else {
		mountDescriptions := strings.Split(strings.TrimSuffix(string(mountDescription), "\n"), " ")
		// mount source, mount target, error
		return mountDescriptions[0], mountDescriptions[1], nil
	}
}
