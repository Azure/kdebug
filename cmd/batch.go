package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/batch"
	"github.com/Azure/kdebug/pkg/formatters"
)

func getBatchDiscoverer(opts *Options, chkCtx *base.CheckContext) batch.BatchDiscoverer {
	if opts.Batch.KubeMachines {
		return batch.NewKubeBatchDiscoverer(chkCtx.KubeClient)
	} else {
		return &batch.StaticBatchDiscoverer{
			Machines: opts.Batch.Machines,
		}
	}
}

func getBatchExecutor(opts *Options) batch.BatchExecutor {
	return &batch.SshBatchExecutor{
		User: opts.Batch.SshUser,
	}
}

func runBatch(opts *Options, chkCtx *base.CheckContext, formatter formatters.Formatter) {
	discoverer := getBatchDiscoverer(opts, chkCtx)
	machines, err := discoverer.Discover()
	if err != nil {
		log.Fatalf("Fail to discover machines: %+v", err)
	}

	executor := getBatchExecutor(opts)
	concurrency := 1
	if opts.Batch.Concurrency > 0 {
		concurrency = opts.Batch.Concurrency
	}
	batchOpts := &batch.BatchOptions{
		Machines:    machines,
		Suites:      opts.Suites,
		Concurrency: concurrency,
	}
	batchResults, err := executor.Execute(batchOpts)
	if err != nil {
		log.Fatal("Fail to run batch", err)
	}

	err = formatter.WriteBatchResults(os.Stdout, batchResults)
	if err != nil {
		log.Fatal(err)
	}
}
