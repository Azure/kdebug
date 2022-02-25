package checker

import (
	"github.com/Azure/kdebug/pkg/checkers/dns"
	"github.com/Azure/kdebug/pkg/checkers/dummy"
	kubeobjectsize "github.com/Azure/kdebug/pkg/checkers/kube/objectsize"
	"github.com/Azure/kdebug/pkg/checkers/oom"
)

var allCheckers = map[string]Checker{
	"dummy":          &dummy.DummyChecker{},
	"dns":            dns.New(),
	"oom":            oom.New(),
	"kubeobjectsize": kubeobjectsize.New(),
}
