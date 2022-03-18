package main

import (
	"fmt"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
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

	Pod struct {
		Name      string `long:"name" description:"Pod name"`
		Namespace string `long:"namespace" description:"Namespace the Pod runs in"`
	} `group:"pod_info" namespace:"pod" description:"Information of a Pod"`
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

	if opts.ListCheckers {
		fmt.Println(chks.ListAllCheckerNames())
		return
	}

	// Prepare dependencies
	ctx, err := buildContext(&opts)
	if err != nil {
		log.Fatal(err)
	}

	// Check
	results, err := chks.Check(ctx, opts.Suites)
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
