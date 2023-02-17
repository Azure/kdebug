package checker

import (
	"sort"

	"github.com/Azure/kdebug/pkg/checkers/diskreadonly"
	"github.com/Azure/kdebug/pkg/checkers/diskusage"
	"github.com/Azure/kdebug/pkg/checkers/dns"
	"github.com/Azure/kdebug/pkg/checkers/dummy"
	"github.com/Azure/kdebug/pkg/checkers/http"
	icmpping "github.com/Azure/kdebug/pkg/checkers/icmp"
	"github.com/Azure/kdebug/pkg/checkers/kmscachesize"
	kubeobjectsize "github.com/Azure/kdebug/pkg/checkers/kube/objectsize"
	"github.com/Azure/kdebug/pkg/checkers/kube/pod"
	"github.com/Azure/kdebug/pkg/checkers/liveness"
	"github.com/Azure/kdebug/pkg/checkers/oom"
	"github.com/Azure/kdebug/pkg/checkers/podschedule"
	"github.com/Azure/kdebug/pkg/checkers/systemload"
	"github.com/Azure/kdebug/pkg/checkers/tcpping"
)

var allCheckers = map[string]Checker{
	"dummy":          &dummy.DummyChecker{},
	"dns":            dns.New(),
	"oom":            oom.New(),
	"kubeobjectsize": kubeobjectsize.New(),
	"diskusage":      diskusage.New(),
	"diskreadonly":   diskreadonly.New(),
	"kubepod":        pod.New(),
	"liveness":       liveness.New(),
	"http":           http.New(),
	"tcp":            tcpping.New(),
	"ping":           icmpping.New(),
	"systemload":     systemload.New(),
	"kmscachesize":   kmscachesize.New(),
	"podschedule":    podschedule.New(),
}

func ListAllCheckerNames() []string {
	names := make([]string, 0, len(allCheckers))
	for n := range allCheckers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
