package main

import (
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
	Suites         []string `short:"s" long:"suite" description:"Check suites"`
	Format         string   `short:"f" long:"format" description:"Output format"`
	KubeMasterUrl  string   `long:"kube-master-url" description:"Kubernetes API server URL"`
	KubeConfigPath string   `long:"kube-config-path" description:"Path to kubeconfig file"`
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

	return ctx, nil
}

func main() {
	// Process options
	var opts Options
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
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
