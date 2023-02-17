package formatters

import (
	"fmt"
	"io"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/batch"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

type TextFormatter struct{}

func (f *TextFormatter) WriteResults(w io.Writer, results []*base.CheckResult) error {
	failures := []*base.CheckResult{}
	for _, r := range results {
		if r.Ok() {
			if log.IsLevelEnabled(log.DebugLevel) {
				fmt.Fprintf(w, "[%s] %s\n", r.Checker, r.Description)
			}
		} else {
			failures = append(failures, r)
		}
	}

	fmt.Fprintf(w, "------------------------------\n")

	if len(failures) == 0 {
		fmt.Fprintf(w, "All %v checks passed!\n",
			color.GreenString("%d", len(results)))
		return nil
	}

	fmt.Fprintf(w, "%v checks passed. %v failed.\n",
		color.GreenString("%d", len(results)-len(failures)),
		color.RedString("%d", len(failures)))
	fmt.Fprintf(w, "------------------------------\n")
	fmt.Fprintf(w, "kdebug has detected these problems for you:\n")

	for _, r := range failures {
		fmt.Fprintf(w, "------------------------------\n")
		fmt.Fprintf(w, color.YellowString("Checker: %s\n", r.Checker))
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

func (f *TextFormatter) WriteBatchResults(w io.Writer, results []*batch.BatchResult) error {
	for _, result := range results {
		fmt.Fprintf(w, color.BlueString("=============== Machine: %s ===============\n",
			result.Machine))
		if result.Error == nil {
			f.WriteResults(w, result.CheckResults)
		} else {
			fmt.Fprintf(w, "Remote execution error: %s\n", result.Error)
		}
	}
	return nil
}
