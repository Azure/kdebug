package base

type CheckContext struct {
	// TODO: Add user input here
	// TODO: Add shared dependencies here, for example, kube-client
}

type CheckResult struct {
	Checker         string
	Error           string
	Description     string
	Recommandations []string
	Logs            []string
	HelpLinks       []string
}

func (r *CheckResult) Ok() bool {
	return r.Error == ""
}
