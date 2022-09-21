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
	Args   []string
	Config interface{}
}

type Netexec struct {
	Pid       string `long:"pid" description:"Attach into a specific pid's network namespace."`
	PodName   string `long:"pod" description:"Attach into a specific pod's network namespace. Caution: The command will use ephemeral debug container to attach a container with 'ghcr.io/azure/kdebug:main' to the target pod."`
	Namespace string `long:"namespace" description:"the namespace of the pod."`
	Command   string `long:"command" description:"Customize the command to be run in container namespace. Leave it blank to use 'sh'."`
	Image     string `long:"image" description:"Customize the image to be used to run command when using --netexec.pod. Leave it blank to use busybox."`
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
