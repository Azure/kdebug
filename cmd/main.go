package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	flags "github.com/jessevdk/go-flags"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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
	Help           bool     `short:"h" long:"help" description:"Show help message"`
	NoSetExitCode  bool     `long:"no-set-exit-code" hidden:"-"`
	Output         string   `short:"o" long:"output" description:"Output file"`

	Batch struct {
		KubeMachines              bool     `long:"kube-machines" description:"Discover machines from Kubernetes API server"`
		KubeMachinesUnready       bool     `long:"kube-machines-unready" description:"Discover unready machines from Kubernetes API server"`
		KubeMachinesLabelSelector string   `long:"kube-machines-label" description:"Label selector for Kubernetes machines"`
		Machines                  []string `long:"machines" description:"Machine names"`
		MachinesFile              string   `long:"machines-file" description:"Path to a file that contains machine names list. Can use - to read from stdin."`
		Concurrency               int      `long:"concurrency" default:"4" description:"Batch concurrency"`
		SshUser                   string   `long:"ssh-user" description:"SSH user"`
		PodExecutorImage          string   `long:"pod-executor-image" description:"Container image used by pod executor" default:"ghcr.io/azure/kdebug:main"`
		PodExecutorNamespace      string   `long:"pod-executor-namespace" description:"Namespace used by pod executor" default:"kdebug"`
		PodExecutorMode           string   `long:"pod-executor-mode" choice:"host" choice:"container" default:"host" description:"Run as container or run as host"`
	} `group:"Batch Options" namespace:"batch" description:"Batch mode"`

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

func buildKubeClient(masterUrl, kubeConfigPath string) (*kubernetes.Clientset, *genericclioptions.ConfigFlags, error) {
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
		return nil, nil, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	kubeConfigFlag := genericclioptions.NewConfigFlags(false)
	kubeConfigFlag.APIServer = &masterUrl
	kubeConfigFlag.KubeConfig = &kubeConfigPath

	return clientSet, kubeConfigFlag, nil
}

func buildCheckContext(opts *Options) (*base.CheckContext, error) {
	ctx := &base.CheckContext{
		Environment: env.GetEnvironment(),
	}

	log.WithFields(log.Fields{
		"env": ctx.Environment,
	}).Debug("Environment")

	kubeClient, _, err := buildKubeClient(opts.KubeMasterUrl, opts.KubeConfigPath)
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
	// Add back help arg so tool can see it
	if opts.Help {
		opts.RemainingArgs = append(opts.RemainingArgs, "-h")
	}
	log.WithFields(log.Fields{"args": opts.RemainingArgs}).Debug("Tool context")
	ctx := &base.ToolContext{
		Args:        opts.RemainingArgs,
		Environment: env.GetEnvironment(),
	}
	if _, configFlags, err := buildKubeClient(opts.KubeMasterUrl, opts.KubeConfigPath); err == nil {
		ctx.KubeConfigFlag = configFlags
	}
	return ctx, nil
}

func main() {
	// Process options
	var opts Options
	flagsParser := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.IgnoreUnknown)
	remainingArgs, err := flagsParser.Parse()
	if err != nil {
		log.Fatal(err)
		return
	}
	opts.RemainingArgs = remainingArgs

	processOptions(&opts)

	if len(opts.Verbose) > 0 {
		if opts.Verbose == "none" {
			logrus.SetOutput(ioutil.Discard)
		} else {
			logLevel, err := logrus.ParseLevel(opts.Verbose)
			if err != nil {
				log.Fatal(err)
			}
			logrus.SetLevel(logLevel)
		}
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) || opts.NoColor || opts.Output != "" {
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

		err = tools.ParseArgs(ctx, opts.Tool, opts.RemainingArgs)
		if err != nil {
			if !flags.WroteHelp(err) {
				log.Fatal(err)
			}
			return
		}

		err = tools.Run(ctx, opts.Tool)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if opts.Help {
		flagsParser.WriteHelp(os.Stdout)
		return
	}

	// Prepare dependencies
	ctx, err := buildCheckContext(&opts)
	if err != nil {
		log.Fatal(err)
	}

	ctx.Output = os.Stdout
	if opts.Output != "" {
		outFile, err := os.OpenFile(opts.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Fail to open output file: %s", opts.Output)
		}
		defer outFile.Close()
		ctx.Output = outFile
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
	err = formatter.WriteResults(ctx.Output, results)
	if err != nil {
		log.Fatal(err)
	}

	if !opts.NoSetExitCode {
		for _, r := range results {
			if !r.Ok() {
				os.Exit(1)
			}
		}
	}
}
