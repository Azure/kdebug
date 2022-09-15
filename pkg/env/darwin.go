//go:build darwin

package env

import "runtime"

func getLinuxFlags() []string {
	return []string{
		runtime.GOOS,
	}
}
