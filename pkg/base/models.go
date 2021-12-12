package base

import "github.com/Azure/kdebug/pkg/env"

type CheckContext struct {
	// TODO: Add user input here
	// TODO: Add shared dependencies here, for example, kube-client
	Environment *env.Environment
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
