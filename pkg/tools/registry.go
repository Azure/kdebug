package tools

import (
	"sort"

	"github.com/nevembalint/kdebug/pkg/tools/aadssh"
	"github.com/nevembalint/kdebug/pkg/tools/netexec"
	"github.com/nevembalint/kdebug/pkg/tools/tcpdump"
	"github.com/nevembalint/kdebug/pkg/tools/upgradeinspector"
	"github.com/nevembalint/kdebug/pkg/tools/vmrebootdetector"
)

var allTools = map[string]Tool{
	"tcpdump":         tcpdump.New(),
	"vmrebootinspect": vmrebootdetector.New(),
	"upgradesinspect": upgradeinspector.New(),
	"aadssh":          aadssh.New(),
	"netexec":         netexec.New(),
}

func ListAllToolNames() []string {
	names := make([]string, 0, len(allTools))
	for n := range allTools {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
