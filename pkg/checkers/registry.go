package checker

import (
	"github.com/Azure/kdebug/pkg/checkers/diskusage"
	"github.com/Azure/kdebug/pkg/checkers/dns"
	"github.com/Azure/kdebug/pkg/checkers/dummy"
	kubeobjectsize "github.com/Azure/kdebug/pkg/checkers/kube/objectsize"
)

var allCheckers = map[string]Checker{
	"dummy":          &dummy.DummyChecker{},
	"dns":            dns.New(),
	"kubeobjectsize": kubeobjectsize.New(),
	"diskusage":      diskusage.New(),
}
