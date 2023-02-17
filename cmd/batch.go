package main

import (
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/batch"
	"github.com/Azure/kdebug/pkg/formatters"
)

func getBatchDiscoverer(opts *Options, chkCtx *base.CheckContext) batch.BatchDiscoverer {
	if opts.Batch.KubeMachines || opts.Batch.KubeMachinesUnready || len(opts.Batch.KubeMachinesLabelSelector) > 0 {
		return batch.NewKubeBatchDiscoverer(chkCtx.KubeClient, opts.Batch.KubeMachinesLabelSelector, opts.Batch.KubeMachinesUnready)
	} else if opts.Batch.MachinesFile != "" {
		return &batch.FileBatchDiscoverer{
			Path: opts.Batch.MachinesFile,
		}
	} else {
		return &batch.StaticBatchDiscoverer{
			Machines: opts.Batch.Machines,
		}
	}
}

func getBatchExecutor(opts *Options, chkCtx *base.CheckContext) batch.BatchExecutor {
	if opts.Batch.SshUser != "" {
		return batch.NewSshBatchExecutor(opts.Batch.SshUser)
	} else if chkCtx.KubeClient != nil {
		return batch.NewPodBatchExecutor(
			chkCtx.KubeClient,
			opts.Batch.PodExecutorImage,
			opts.Batch.PodExecutorNamespace,
		)
	} else {
		log.Fatal("No batch executor configured")
		return nil
	}
}

type batchReporter struct {
	out io.Writer
	bar *progressbar.ProgressBar
}

func newBatchReporter(out io.Writer, max int64) *batchReporter {
	return &batchReporter{
		out: out,
		bar: progressbar.Default(max),
	}
}

func (r *batchReporter) OnResult(result *batch.BatchResult) {
	r.bar.Add(1)
}

func runBatch(opts *Options, chkCtx *base.CheckContext, formatter formatters.Formatter) {
	discoverer := getBatchDiscoverer(opts, chkCtx)
	machines, err := discoverer.Discover()
	if err != nil {
		log.Fatalf("Fail to discover machines: %+v", err)
	}

	log.WithFields(log.Fields{"count": len(machines)}).Info("Discovered machines list")

	executor := getBatchExecutor(opts, chkCtx)
	concurrency := 1
	if opts.Batch.Concurrency > 0 {
		concurrency = opts.Batch.Concurrency
	}
	batchOpts := &batch.BatchOptions{
		Machines:    machines,
		Checkers:    opts.Checkers,
		Concurrency: concurrency,
		Reporter:    newBatchReporter(os.Stdout, int64(len(machines))),
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
