package checker

import (
	"github.com/Azure/kdebug/pkg/checkers/dns"
	"github.com/Azure/kdebug/pkg/checkers/dummy"
)

var allCheckers = map[string]Checker{
	"dummy": &dummy.DummyChecker{},
	"dns":   dns.New(),
}