package base

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
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
	Args           []string
	Config         interface{}
	KubeConfigFlag *genericclioptions.ConfigFlags
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
