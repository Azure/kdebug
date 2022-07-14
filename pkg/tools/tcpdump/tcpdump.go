package tcpdump

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
	log "github.com/sirupsen/logrus"
)

type TcpdumpTool struct {
	srcIP    string
	srcPort  string
	dstIP    string
	dstPort  string
	hostIP   string
	hostPort string
	pid      string
	tcpOnly  bool
}

const (
	DefaultTcpdumpArguments = "-nvvv"
)

func New() *TcpdumpTool {
	return &TcpdumpTool{}
}

func (c *TcpdumpTool) Name() string {
	return "Tcpdump"
}

func logAndExec(name string, args ...string) *exec.Cmd {
	log.Infof("Exec %s %+v", name, args)
	return exec.Command(name, args...)
}

func (c *TcpdumpTool) Run(ctx *base.ToolContext) error {
	c.ParseParameters(ctx)
	tcpdumpArgs := c.GenerateTcpdumpParamerters()

	// Attch pid
	if len(ctx.Tcpdump.Pid) > 0 {
		_, err := logAndExec("nsenter", "-n", "-t", ctx.Tcpdump.Pid).Output()

		if err != nil {
			return err
		}
	}

	cmd := logAndExec("tcpdump", strings.Split(tcpdumpArgs, " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func (c *TcpdumpTool) ParseParameters(ctx *base.ToolContext) {
	c.srcIP, c.srcPort = ParseIPAndPort(ctx.Tcpdump.Source)
	c.dstIP, c.dstPort = ParseIPAndPort(ctx.Tcpdump.Destination)
	c.hostIP, c.hostPort = ParseIPAndPort(ctx.Tcpdump.Host)
	c.pid = ctx.Tcpdump.Pid
	c.tcpOnly = ctx.Tcpdump.TcpOnly
}

func (c *TcpdumpTool) GenerateTcpdumpParamerters() string {
	var cmd []string
	if len(c.srcIP) > 0 {
		cmd = append(cmd, fmt.Sprintf("src %s", c.srcIP))
	}
	if len(c.srcPort) > 0 {
		cmd = append(cmd, fmt.Sprintf("src port %s", c.srcPort))
	}
	if len(c.dstIP) > 0 {
		cmd = append(cmd, fmt.Sprintf("dst %s", c.dstIP))
	}
	if len(c.dstPort) > 0 {
		cmd = append(cmd, fmt.Sprintf("dst port %s", c.dstPort))
	}
	if len(c.hostIP) > 0 {
		cmd = append(cmd, fmt.Sprintf("host %s", c.hostIP))
	}
	if len(c.hostPort) > 0 {
		cmd = append(cmd, fmt.Sprintf("port %s", c.hostPort))
	}
	if c.tcpOnly {
		cmd = append(cmd, "tcp")
	}
	return DefaultTcpdumpArguments + " " + strings.Join(cmd, " and ")
}

func ParseIPAndPort(s string) (ip string, port string) {
	colon := strings.Index(s, ":")
	if colon == -1 {
		return s, ""
	}

	return s[0:colon], s[colon+1:]
}
