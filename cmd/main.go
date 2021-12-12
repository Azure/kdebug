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
	Format string   `short:"f" long:"format" description:"Output format"`
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
	var formatter formatters.Formatter
	if opts.Format == "json" {
		formatter = &formatters.JsonFormatter{}
	} else {
		formatter = &formatters.TextFormatter{}
	}

	err = formatter.WriteResults(os.Stdout, results)
	if err != nil {
		log.Fatal(err)
	}
}
