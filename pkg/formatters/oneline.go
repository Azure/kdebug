package formatters

import (
	"fmt"
	"io"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/batch"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

type OneLineFormatter struct{}

func (f *OneLineFormatter) WriteResults(w io.Writer, results []*base.CheckResult) error {
	failedCheckers := make(map[string]struct{})
	failures := []*base.CheckResult{}
	for _, r := range results {
		if r.Ok() {
			if log.IsLevelEnabled(log.DebugLevel) {
				fmt.Fprintf(w, "[%s] %s\n", r.Checker, r.Description)
			}
		} else {
			failures = append(failures, r)
			failedCheckers[r.Checker] = struct{}{}
		}
	}

	if len(failures) == 0 {
		fmt.Fprintf(w, "All %v checks passed!\n",
			color.GreenString("%d", len(results)))
		return nil
	}

	failedCheckersList := []string{}
	for c := range failedCheckers {
		failedCheckersList = append(failedCheckersList, c)
	}

	fmt.Fprintf(w, "%v checks checked, %v failed: %s",
		color.GreenString("%d", len(results)),
		color.RedString("%d", len(failures)),
		strings.Join(failedCheckersList, ", "))

	return nil
}

func (f *OneLineFormatter) WriteBatchResults(w io.Writer, results []*batch.BatchResult) error {
	return fmt.Errorf("not implemented: one line formatter for batch results")
}
