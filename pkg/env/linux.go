//go:build linux

package env

import (
	"runtime"
	"strings"

	"github.com/zcalusic/sysinfo"
)

func getLinuxFlags() []string {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	return []string{
		runtime.GOOS,
		strings.ToLower(si.OS.Vendor),
	}
}
