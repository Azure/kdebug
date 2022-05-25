package batch

import "github.com/Azure/kdebug/pkg/base"

type BatchOptions struct {
	Machines    []string
	Suites      []string
	Concurrency int
	Reporter    BatchReportor
}

type batchTask struct {
	Machine string
	Suites  []string
}

type BatchResult struct {
	Machine      string
	Error        error
	CheckResults []*base.CheckResult
}

type BatchExecutor interface {
	Execute(opts *BatchOptions) ([]*BatchResult, error)
}

type BatchReportor interface {
	OnResult(result *BatchResult)
}

type BatchDiscoverer interface {
	Discover() ([]string, error)
}
