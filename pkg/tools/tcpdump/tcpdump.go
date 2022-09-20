package tcpdump

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	flags "github.com/jessevdk/go-flags"

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

type Config struct {
	Source      string `long:"source" description:"The source of the connection. Format: <ip>:<port>. Watch all sources if not assigned."`
	Destination string `long:"destination" description:"The destination of the connection. Format: <ip>:<port>. Watch all destination if not assigned."`
	Host        string `long:"host" description:"The host(either src or dst) of the connection. Format: <ip>:<port>. Watch if not assigned."`
	Pid         string `short:"p" long:"pid" description:"Attach into a specific pid's network namespace. Use current namespace if not assigned"`
	TcpOnly     bool   `long:"tcponly" description:"Only watch tcp connections"`
}

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

func (c *TcpdumpTool) ParseArgs(ctx *base.ToolContext, args []string) error {
	var config Config
	remainingArgs, err := flags.ParseArgs(&config, args)
	if err != nil {
		return err
	}
	ctx.Config = &config
	ctx.Args = remainingArgs
	return nil
}

func (c *TcpdumpTool) Run(ctx *base.ToolContext) error {
	config := ctx.Config.(*Config)
	c.ParseParameters(config)
	tcpdumpArgs := c.GenerateTcpdumpParamerters()

	// Attch pid
	if len(config.Pid) > 0 {
		_, err := logAndExec("nsenter", "-n", "-t", config.Pid).Output()

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

func (c *TcpdumpTool) ParseParameters(config *Config) {
	c.srcIP, c.srcPort = ParseIPAndPort(config.Source)
	c.dstIP, c.dstPort = ParseIPAndPort(config.Destination)
	c.hostIP, c.hostPort = ParseIPAndPort(config.Host)
	c.pid = config.Pid
	c.tcpOnly = config.TcpOnly
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
