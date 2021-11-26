package formatters

import (
	"io"

	"github.com/Azure/kdebug/pkg/base"
)

type Formatter interface {
	Format(io.Writer, []*base.CheckResult) error
}
