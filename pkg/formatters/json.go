package formatters

import (
	"encoding/json"
	"io"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/batch"
)

type JsonFormatter struct{}

func (f *JsonFormatter) WriteResults(w io.Writer, results []*base.CheckResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(results)
}

func (f *JsonFormatter) WriteBatchResults(w io.Writer, results []*batch.BatchResult) error {
	// TODO
	return nil
}
