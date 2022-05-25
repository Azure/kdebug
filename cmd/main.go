package main

import (
	"fmt"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/Azure/kdebug/pkg/base"
	chks "github.com/Azure/kdebug/pkg/checkers"
	"github.com/Azure/kdebug/pkg/env"
	"github.com/Azure/kdebug/pkg/formatters"
)

type Options struct {
	ListCheckers   bool     `short:"l" long:"list" description:"List all checkers"`
	Suites         []string `short:"s" long:"suite" description:"Check suites"`
	Format         string   `short:"f" long:"format" description:"Output format"`
	KubeMasterUrl  string   `long:"kube-master-url" description:"Kubernetes API server URL"`
	KubeConfigPath string   `long:"kube-config-path" description:"Path to kubeconfig file"`
	Verbose        string   `short:"v" long:"verbose" description:"Log level"`

	Pod struct {
		Name      string `long:"name" description:"Pod name"`
		Namespace string `long:"namespace" description:"Namespace the Pod runs in"`
	} `group:"pod_info" namespace:"pod" description:"Information of a Pod"`

	Batch struct {
		KubeMachines              bool     `long:"kube-machines" description:"Discover machines from Kubernetes API server"`
		KubeMachinesUnready       bool     `long:"kube-machines-unready" description:"Discover unready machines from Kubernetes API server"`
		KubeMachinesLabelSelector string   `long:"kube-machines-label" description:"Label selector for Kubernetes machines"`
		Machines                  []string `long:"machines" description:"Machine names"`
		MachinesFile              string   `long:"machines-file" description:"Path to a file that contains machine names list. Can use - to read from stdin."`
		Concurrency               int      `long:"concurrency" default:"4" description:"Batch concurrency"`
		SshUser                   string   `long:"ssh-user" description:"SSH user"`
	} `group:"batch" namespace:"batch" description:"Batch mode"`
}

func (o *Options) IsBatchMode() bool {
	return o.Batch.KubeMachines || o.Batch.KubeMachinesUnready || len(o.Batch.Machines) > 0 || len(o.Batch.MachinesFile) > 0
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

func buildContext(opts *Options) (*base.CheckContext, error) {
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

	ctx.Pod = struct {
		Name      string
		Namespace string
	}(opts.Pod)

	return ctx, nil
}

func main() {
	// Process options
	var opts Options
	_, err := flags.Parse(&opts)
	if err != nil {
		if !flags.WroteHelp(err) {
			log.Fatal(err)
		}
		return
	}

	if len(opts.Verbose) > 0 {
		logLevel, err := logrus.ParseLevel(opts.Verbose)
		if err != nil {
			log.Fatal(err)
		}
		logrus.SetLevel(logLevel)
	}

	if opts.ListCheckers {
		fmt.Println(chks.ListAllCheckerNames())
		return
	}

	var formatter formatters.Formatter
	if opts.Format == "json" {
		formatter = &formatters.JsonFormatter{}
	} else {
		formatter = &formatters.TextFormatter{}
	}

	// Prepare dependencies
	ctx, err := buildContext(&opts)
	if err != nil {
		log.Fatal(err)
	}

	// Batch mode
	if opts.IsBatchMode() {
		runBatch(&opts, ctx, formatter)
		return
	}

	// Check
	results, err := chks.Check(ctx, opts.Suites)
	if err != nil {
		log.Fatal(err)
	}

	// Output
	err = formatter.WriteResults(os.Stdout, results)
	if err != nil {
		log.Fatal(err)
	}
}
