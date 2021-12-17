package formatters

import (
	"fmt"
	"io"

	"github.com/Azure/kdebug/pkg/base"
)

type TextFormatter struct{}

func (f *TextFormatter) WriteResults(w io.Writer, results []*base.CheckResult) error {
	failures := []*base.CheckResult{}
	for _, r := range results {
		if r.Ok() {
			fmt.Fprintf(w, "[%s] %s\n", r.Checker, r.Description)
		} else {
			failures = append(failures, r)
		}
	}

	fmt.Fprintf(w, "------------------------------\n")

	if len(failures) == 0 {
		fmt.Fprintf(w, "All %d checks passed!\n", len(results))
		return nil
	}

	fmt.Fprintf(w, "kdebug has detected these problems for you:\n")

	for _, r := range failures {
		fmt.Fprintf(w, "------------------------------\n")
		fmt.Fprintf(w, "Checker: %s\n", r.Checker)
		fmt.Fprintf(w, "Error: %s\n", r.Error)
		fmt.Fprintf(w, "Description: %s\n", r.Description)
		if len(r.Recommendations) > 0 {
			fmt.Fprintf(w, "Recommendations:\n")
			for i, rec := range r.Recommendations {
				fmt.Fprintf(w, "[%d] %s\n", i+1, rec)
			}
		}
		// TODO: Make logs more pretty
		if len(r.Logs) > 0 {
			fmt.Fprintf(w, "Logs:\n")
			for _, l := range r.Logs {
				fmt.Fprintf(w, "%s\n", l)
			}
		}
		if len(r.HelpLinks) > 0 {
			fmt.Fprintf(w, "Help links:\n")
			for i, l := range r.HelpLinks {
				fmt.Fprintf(w, "[%d] %s\n", i+1, l)
			}
		}
	}

	return nil
}
