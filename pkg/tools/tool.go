package tools

import (
	"errors"

	"github.com/Azure/kdebug/pkg/base"
)

type Tool interface {
	Name() string
	Run(*base.CheckContext) error
}

func Run(ctx *base.CheckContext, suite string) error {
	if tool, ok := allTools[suite]; ok {
		err := tool.Run(ctx)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Unknown checker: " + suite)
	}

	return nil
}
