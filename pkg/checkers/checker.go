package checker

import (
	"errors"
	"log"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
)

type Checker interface {
	Name() string
	Check(*base.CheckContext) ([]*base.CheckResult, error)
}

func Check(suites []string) ([]*base.CheckResult, error) {
	ctx := &base.CheckContext{
		Environment: env.GetEnvironment(),
	}

	checkers := make([]Checker, 0, len(suites))
	for _, suite := range suites {
		if checker, ok := allCheckers[suite]; ok {
			checkers = append(checkers, checker)
		} else {
			return nil, errors.New("Unknown checker: " + suite)
		}
	}

	var results []*base.CheckResult
	for _, checker := range checkers {
		r, err := checker.Check(ctx)
		if err != nil {
			log.Printf("Error in checker %s: %s", err, checker.Name())
		}
		results = append(results, r...)
	}

	return results, nil
}