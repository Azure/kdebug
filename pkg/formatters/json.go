package formatters

import (
	"encoding/json"
	"io"

	"github.com/Azure/kdebug/pkg/base"
)

type JsonFormatter struct{}

func (f *JsonFormatter) WriteResults(w io.Writer, results []*base.CheckResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(results)
}
