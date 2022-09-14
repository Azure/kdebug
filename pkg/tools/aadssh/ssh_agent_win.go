//go:build windows

package aadssh

import (
	"net"

	"github.com/Microsoft/go-winio"
)

func dialSSHAgent(path string) (net.Conn, error) {
	return winio.DialPipe(path, nil)
}
