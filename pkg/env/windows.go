//go:build windows

package env

import (
	"runtime"
)

func getLinuxFlags() []string {
	return []string{
		runtime.GOOS,
	}
}
