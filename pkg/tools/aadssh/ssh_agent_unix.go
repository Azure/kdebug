//go:build !windows

package aadssh

import "net"

func dialSSHAgent(path string) (net.Conn, error) {
	return net.Dial("unix", path)
}
