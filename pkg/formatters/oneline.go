package formatters

import (
	"fmt"
	"io"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/batch"
	"github.com/fatih/color"
)

type OnelineFormatter struct{}

func (f *OnelineFormatter) WriteResults(w io.Writer, results []*base.CheckResult) error {
	failures := []*base.CheckResult{}
	for _, r := range results {
		if !r.Ok() {
			failures = append(failures, r)
		}
	}

	if len(failures) == 0 {
		fmt.Fprintf(w, "All %v checks passed!",
			color.GreenString("%d", len(results)))
		return nil
	}

	for _, r := range failures {
		fmt.Fprintf(w, color.YellowString("[%s] ", r.Checker))
		fmt.Fprintf(w, "%s ", r.Error)
	}

	return nil
}

func (f *OnelineFormatter) WriteBatchResults(w io.Writer, results []*batch.BatchResult) error {
	for _, result := range results {
		fmt.Fprintf(w, color.BlueString("[%s] ",
			result.Machine))
		if result.Error == nil {
			f.WriteResults(w, result.CheckResults)
		} else {
			fmt.Fprintf(w, "Remote execution error: %s ", result.Error)
		}
	}
	return nil
}
