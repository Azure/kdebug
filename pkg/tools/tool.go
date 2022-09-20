package tools

import (
	"errors"

	"github.com/Azure/kdebug/pkg/base"
)

type Tool interface {
	Name() string
	ParseArgs(*base.ToolContext, []string) error
	Run(*base.ToolContext) error
}

func getTool(name string) (Tool, error) {
	if tool, ok := allTools[name]; ok {
		return tool, nil
	} else {
		return nil, errors.New("Unknown tool: " + name)
	}
}

func ParseArgs(ctx *base.ToolContext, name string, args []string) error {
	tool, err := getTool(name)
	if err != nil {
		return err
	}
	return tool.ParseArgs(ctx, args)
}

func Run(ctx *base.ToolContext, name string) error {
	tool, err := getTool(name)
	if err != nil {
		return err
	}
	return tool.Run(ctx)
}
