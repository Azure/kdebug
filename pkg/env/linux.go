package env

import (
	"strings"

	"github.com/zcalusic/sysinfo"
)

func getLinuxFlags() []string {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	return []string{
		strings.ToLower(si.OS.Vendor),
	}
}
