package tools

import (
	"sort"

	"github.com/Azure/kdebug/pkg/tools/aadssh"
	"github.com/Azure/kdebug/pkg/tools/tcpdump"
	"github.com/Azure/kdebug/pkg/tools/upgradeinspector"
	"github.com/Azure/kdebug/pkg/tools/vmrebootdetector"
)

var allTools = map[string]Tool{
	"tcpdump":          tcpdump.New(),
	"vmrebootdetector": vmrebootdetector.New(),
	"upgradeinspector": upgradeinspector.New(),
	"aadssh":           aadssh.New(),
}

func ListAllToolNames() []string {
	names := make([]string, 0, len(allTools))
	for n := range allTools {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
