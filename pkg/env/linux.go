//go:build linux

package env

import (
	"os"
	"runtime"
	"strings"

	"github.com/zcalusic/sysinfo"
)

func getLinuxFlags() []string {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	flags := []string{
		runtime.GOOS,
		strings.ToLower(si.OS.Vendor),
	}

	if os.Geteuid() == 0 {
		flags = append(flags, "root")
	}

	return flags
}
