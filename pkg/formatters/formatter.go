package formatters

import (
	"io"

	"github.com/Azure/kdebug/pkg/base"
)

type Formatter interface {
	WriteResults(io.Writer, []*base.CheckResult) error
}
