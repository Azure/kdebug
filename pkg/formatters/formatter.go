package formatters

import (
	"io"

	"github.com/nevembalint/kdebug/pkg/base"
	"github.com/nevembalint/kdebug/pkg/batch"
)

type Formatter interface {
	WriteResults(io.Writer, []*base.CheckResult) error
	WriteBatchResults(io.Writer, []*batch.BatchResult) error
}
