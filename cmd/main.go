package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	flags "github.com/jessevdk/go-flags"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/Azure/kdebug/pkg/base"
	chks "github.com/Azure/kdebug/pkg/checkers"
	"github.com/Azure/kdebug/pkg/env"
	"github.com/Azure/kdebug/pkg/formatters"
	tools "github.com/Azure/kdebug/pkg/tools"
)

type Options struct {
	ListCheckers   bool     `short:"l" long:"list" description:"List all checks and tools"`
	Checkers       []string `short:"c" long:"check" description:"Check name. Can specify multiple times."`
	Tool           string   `short:"t" long:"tool" description:"Use tool"`
	Format         string   `short:"f" long:"format" description:"Output format"`
	KubeMasterUrl  string   `long:"kube-master-url" description:"Kubernetes API server URL"`
	KubeConfigPath string   `long:"kube-config-path" description:"Path to kubeconfig file"`
	Verbose        string   `short:"v" long:"verbose" description:"Log level"`
	NoColor        bool     `long:"no-color" description:"Disable colorized output"`
	Pause          bool     `long:"pause" description:"Pause until interrupted"`

	Batch struct {
		KubeMachines              bool     `long:"kube-machines" description:"Discover machines from Kubernetes API server"`
		KubeMachinesUnready       bool     `long:"kube-machines-unready" description:"Discover unready machines from Kubernetes API server"`
		KubeMachinesLabelSelector string   `long:"kube-machines-label" description:"Label selector for Kubernetes machines"`
		Machines                  []string `long:"machines" description:"Machine names"`
		MachinesFile              string   `long:"machines-file" description:"Path to a file that contains machine names list. Can use - to read from stdin."`
		Concurrency               int      `long:"concurrency" default:"4" description:"Batch concurrency"`
		SshUser                   string   `long:"ssh-user" description:"SSH user"`
	} `group:"batch" namespace:"batch" description:"Batch mode"`

	Tcpdump        base.Tcpdump          `group:"tcpdump" namespace:"tcpdump" description:"Tool mode: tcpdump"`
	VMRebootDetect base.VMRebootDetector `group:"vmrebootdetector" namespace:"vmrebootdetector" description:"Tool mode: vm reboot detector"`
	UpgradeInspect base.UpgradeInspector `group:"upgradeinspector" namespace:"upgradeinspector" description:"Tool mode: pkg upgrade inspector"`
	AadSsh         base.AadSsh           `group:"aadssh" namespace:"aadssh" description:"Tool mode: AAD SSH"`

	RemainingArgs []string
}

func (o *Options) IsBatchMode() bool {
	return o.Batch.KubeMachines || o.Batch.KubeMachinesUnready || len(o.Batch.Machines) > 0 || len(o.Batch.MachinesFile) > 0
}

func (o *Options) IsToolMode() bool {
	return len(o.Tool) > 0
}

func processOptions(o *Options) {
	// Run all checkers if not specified
	if len(o.Checkers) == 0 {
		o.Checkers = chks.ListAllCheckerNames()
	}
}

func buildKubeClient(masterUrl, kubeConfigPath string) (*kubernetes.Clientset, error) {
	// Try env
	if kubeConfigPath == "" {
		if path := os.Getenv("KUBECONFIG"); path != "" {
			kubeConfigPath = path
		}
	}

	// Try default path
	if kubeConfigPath == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeConfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func buildCheckContext(opts *Options) (*base.CheckContext, error) {
	ctx := &base.CheckContext{
		Environment: env.GetEnvironment(),
	}

	kubeClient, err := buildKubeClient(opts.KubeMasterUrl, opts.KubeConfigPath)
	if err == nil {
		ctx.KubeClient = kubeClient
	} else {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Kubernetes related checkers will not work")
	}

	return ctx, nil
}

func buildToolContext(opts *Options) (*base.ToolContext, error) {
	ctx := &base.ToolContext{
		Args: opts.RemainingArgs,
	}
	ctx.Tcpdump = opts.Tcpdump
	ctx.VmRebootDetector = opts.VMRebootDetect
	ctx.UpgradeInspector = opts.UpgradeInspect
	ctx.AadSsh = opts.AadSsh

	return ctx, nil
}

func main() {
	// Process options
	var opts Options
	remainingArgs, err := flags.Parse(&opts)
	if err != nil {
		if !flags.WroteHelp(err) {
			log.Fatal(err)
		}
		return
	}
	opts.RemainingArgs = remainingArgs

	processOptions(&opts)

	if len(opts.Verbose) > 0 {
		logLevel, err := logrus.ParseLevel(opts.Verbose)
		if err != nil {
			log.Fatal(err)
		}
		logrus.SetLevel(logLevel)
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) || opts.NoColor {
		color.NoColor = true
	}

	if opts.ListCheckers {
		fmt.Print("checks: ")
		fmt.Println(chks.ListAllCheckerNames())
		fmt.Print("tools: ")
		fmt.Println(tools.ListAllToolNames())
		return
	}

	if opts.Pause {
		pause()
		return
	}

	var formatter formatters.Formatter
	if opts.Format == "json" {
		formatter = &formatters.JsonFormatter{}
	} else if opts.Format == "oneline" {
		formatter = &formatters.OneLineFormatter{}
	} else {
		formatter = &formatters.TextFormatter{}
	}

	// Tool Mode
	if opts.IsToolMode() {
		ctx, err := buildToolContext(&opts)
		if err != nil {
			log.Fatal(err)
		}

		err = tools.Run(ctx, opts.Tool)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Prepare dependencies
	ctx, err := buildCheckContext(&opts)
	if err != nil {
		log.Fatal(err)
	}

	// Batch mode
	if opts.IsBatchMode() {
		runBatch(&opts, ctx, formatter)
		return
	}

	// Check
	results, err := chks.Check(ctx, opts.Checkers)
	if err != nil {
		log.Fatal(err)
	}

	// Output
	err = formatter.WriteResults(os.Stdout, results)
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range results {
		if !r.Ok() {
			os.Exit(1)
		}
	}
}
