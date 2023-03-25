package main

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	log "github.com/sirupsen/logrus"
)

const (
	SystemdConfigDir    = "/etc/systemd/system"
	SystemdUnitName     = "kdebug.service"
	SystemdUnitTemplate = `[Unit]
Description=kdebug

[Service]
Type=oneshot
ExecStart=TODO_EXEC_START
TimeoutSec=60

[Install]
WantedBy=multi-user.target
`
	OutputFile = "/tmp/kdebug.stdout.log"
)

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

func writeSystemdUnit(cmd string) error {
	unitConfig := strings.Replace(SystemdUnitTemplate,
		"TODO_EXEC_START", cmd, 1)
	unitConfigPath := path.Join(SystemdConfigDir, SystemdUnitName)
	return ioutil.WriteFile(unitConfigPath, []byte(unitConfig), 0644)
}

func removeSystemdUnit() error {
	unitConfigPath := path.Join(SystemdConfigDir, SystemdUnitName)
	return os.Remove(unitConfigPath)
}

func readOutputs() ([]byte, error) {
	f, err := os.Open(OutputFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("not enough args")
	}

	cmd := os.Args[1]
	cmdArgs := append(os.Args[2:], "--output", OutputFile)

	// Copy binary to host
	baseName := path.Base(cmd)
	dstPath := path.Join("/tmp", baseName)
	if err := copyFile(cmd, dstPath); err != nil {
		log.Fatalf("fail to copy file: %+v", err)
	}

	// Set up system config
	dstCmd := dstPath + " " + strings.Join(cmdArgs, " ")
	if err := writeSystemdUnit(dstCmd); err != nil {
		log.Fatalf("fail to write unit file: %+v", err)
	}

	// Invoke
	conn, err := dbus.NewSystemConnectionContext(context.Background())
	if err != nil {
		log.Fatalf("fail to connect to systemd: %+v", err)
	}
	defer conn.Close()

	if err = conn.ReloadContext(context.Background()); err != nil {
		log.Fatalf("fail to reload systemd: %+v", err)
	}

	ch := make(chan string)
	_, err = conn.StartUnitContext(context.Background(),
		SystemdUnitName, "replace", ch)
	if err != nil {
		log.Fatalf("fail to start systemd unit: %+v", err)
	}

	select {
	case <-ch:
		break
	case <-time.After(75 * time.Second):
		log.Fatalf("timeout starting systemd unit")
	}

	output, err := readOutputs()
	if err != nil {
		log.Fatalf("fail to read output: %+v", err)
	}

	// Cleanup
	if err = removeSystemdUnit(); err != nil {
		log.Fatalf("fail to remove systemd unit: %+v", err)
	}

	if err = os.Remove(OutputFile); err != nil {
		log.Fatalf("fail to remove stdout file: %+v", err)
	}

	// Output
	os.Stdout.Write(output)
}
