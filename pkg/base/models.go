package base

import (
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/kdebug/pkg/env"
)

type CheckContext struct {
	// TODO: Add user input here
	Pod struct {
		Name      string
		Namespace string
	}

	// TODO: Add shared dependencies here, for example, kube-client
	Environment env.Environment
	KubeClient  *kubernetes.Clientset
}

type ToolContext struct {
	Tcpdump Tcpdump
}

type Tcpdump struct {
	Source      string `long:"source" description:"The source of the connection. Format: <ip>:<port>. Watch all sources if not assigned."`
	Destination string `long:"destination" description:"The destination of the connection. Format: <ip>:<port>. Watch all destination if not assigned."`
	Host        string `long:"host" description:"The host(either src or dst) of the connection. Format: <ip>:<port>. Watch if not assigned."`
	Pid         string `short:"p" long:"pid" description:"Attach into a specific pid's network namespace. Use current namespace if not assigned"`
	TcpOnly     bool   `long:"tcponly" description:"Only watch tcp connections"`
}

type CheckResult struct {
	Checker         string
	Error           string
	Description     string
	Recommendations []string
	Logs            []string
	HelpLinks       []string
}

func (r *CheckResult) Ok() bool {
	return r.Error == ""
}
