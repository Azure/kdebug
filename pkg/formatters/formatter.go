package formatters

import (
	"io"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/batch"
)

type Formatter interface {
	WriteResults(io.Writer, []*base.CheckResult) error
	WriteBatchResults(io.Writer, []*batch.BatchResult) error
}
