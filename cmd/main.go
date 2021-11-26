package main

import (
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"

	chks "github.com/Azure/kdebug/pkg/checkers"
	"github.com/Azure/kdebug/pkg/formatters"
)

type Options struct {
	Suites []string `short:"s" long:"suite" description:"Check suites"`
}

func main() {
	// Process options
	var opts Options
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	// Check
	results, err := chks.Check(opts.Suites)
	if err != nil {
		log.Fatal(err)
	}

	// Output
	var formatter formatters.JsonFormatter
	err = formatter.Format(os.Stdout, results)
	if err != nil {
		log.Fatal(err)
	}
}
