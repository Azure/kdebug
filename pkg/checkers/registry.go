package checker

import (
	"sort"

	"github.com/Azure/kdebug/pkg/checkers/diskusage"
	"github.com/Azure/kdebug/pkg/checkers/dns"
	"github.com/Azure/kdebug/pkg/checkers/dummy"
	"github.com/Azure/kdebug/pkg/checkers/http"
	kubeobjectsize "github.com/Azure/kdebug/pkg/checkers/kube/objectsize"
	"github.com/Azure/kdebug/pkg/checkers/kube/pod"
	"github.com/Azure/kdebug/pkg/checkers/oom"
)

var allCheckers = map[string]Checker{
	"dummy":          &dummy.DummyChecker{},
	"dns":            dns.New(),
	"oom":            oom.New(),
	"kubeobjectsize": kubeobjectsize.New(),
	"diskusage":      diskusage.New(),
	"kubepod":        pod.New(),
	"http":           http.New(),
}

func ListAllCheckerNames() []string {
	names := make([]string, 0, len(allCheckers))
	for n := range allCheckers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
