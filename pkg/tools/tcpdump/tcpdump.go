package tcpdump

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
)

type TcpDumpTool struct {
	srcIP    string
	srcPort  string
	dstIP    string
	dstPort  string
	hostIP   string
	hostPort string
	pid      string
	tcpOnly  bool
}

func New() *TcpDumpTool {
	return &TcpDumpTool{}
}

func (c *TcpDumpTool) Name() string {
	return "TcpDump"
}

func (c *TcpDumpTool) Run(ctx *base.CheckContext) error {
	c.ParseParameters(ctx)
	tcpdumpArgs := c.GenerateTcpdumpParamerters()

	// Attch pid
	if len(ctx.Tcpdump.Pid) > 0 {
		_, err := exec.Command("nsenter", "-n", "-t", ctx.Tcpdump.Pid).Output()

		if err != nil {
			return err
		}
	}

	cmd := exec.Command("tcpdump", strings.Split(tcpdumpArgs, " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func (c *TcpDumpTool) ParseParameters(ctx *base.CheckContext) {
	c.srcIP, c.srcPort = ParseIPAndPort(ctx.Tcpdump.Source)
	c.dstIP, c.dstPort = ParseIPAndPort(ctx.Tcpdump.Destination)
	c.hostIP, c.hostPort = ParseIPAndPort(ctx.Tcpdump.Host)
	c.pid = ctx.Tcpdump.Pid
	c.tcpOnly = ctx.Tcpdump.TcpOnly
}

func (c *TcpDumpTool) GenerateTcpdumpParamerters() string {
	cmd := "-nvvv "
	firstCmd := true
	if len(c.srcIP) > 0 {
		cmd += AppendCmd(fmt.Sprintf("src %s", c.srcIP), &firstCmd)
	}
	if len(c.srcPort) > 0 {
		cmd += AppendCmd(fmt.Sprintf("src port %s", c.srcPort), &firstCmd)
	}
	if len(c.dstIP) > 0 {
		cmd += AppendCmd(fmt.Sprintf("dst %s", c.dstIP), &firstCmd)
	}
	if len(c.dstPort) > 0 {
		cmd += AppendCmd(fmt.Sprintf("dst port %s", c.dstPort), &firstCmd)
	}
	if len(c.hostIP) > 0 {
		cmd += AppendCmd(fmt.Sprintf("host %s", c.hostIP), &firstCmd)
	}
	if len(c.hostPort) > 0 {
		cmd += AppendCmd(fmt.Sprintf("port %s", c.hostPort), &firstCmd)
	}
	if c.tcpOnly {
		cmd += AppendCmd("tcp", &firstCmd)
	}
	return cmd
}

func AppendCmd(s string, firstCmd *bool) string {
	if *firstCmd {
		*firstCmd = false
		return s
	}
	return fmt.Sprintf(" and %s", s)
}

func ParseIPAndPort(s string) (ip string, port string) {
	colon := strings.Index(s, ":")
	if colon == -1 {
		return s, ""
	}

	return s[0:colon], s[colon+1:]
}
