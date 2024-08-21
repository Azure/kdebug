package checker

import (
	"sort"

	"github.com/nevembalint/kdebug/pkg/checkers/diskreadonly"
	"github.com/nevembalint/kdebug/pkg/checkers/diskusage"
	"github.com/nevembalint/kdebug/pkg/checkers/dns"
	"github.com/nevembalint/kdebug/pkg/checkers/dummy"
	"github.com/nevembalint/kdebug/pkg/checkers/http"
	icmpping "github.com/nevembalint/kdebug/pkg/checkers/icmp"
	"github.com/nevembalint/kdebug/pkg/checkers/kmscachesize"
	kubeobjectsize "github.com/nevembalint/kdebug/pkg/checkers/kube/objectsize"
	"github.com/nevembalint/kdebug/pkg/checkers/kube/pod"
	"github.com/nevembalint/kdebug/pkg/checkers/liveness"
	"github.com/nevembalint/kdebug/pkg/checkers/oom"
	"github.com/nevembalint/kdebug/pkg/checkers/podschedule"
	"github.com/nevembalint/kdebug/pkg/checkers/systemload"
	"github.com/nevembalint/kdebug/pkg/checkers/tcpping"
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
