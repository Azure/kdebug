package netexec

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kubecmd "k8s.io/kubectl/pkg/cmd"
)

type Config struct {
	Pid       string `long:"pid" description:"Attach into a specific pid's network namespace."`
	PodName   string `long:"pod" description:"Attach into a specific pod's network namespace. Caution: The command will use ephemeral debug container to attach a container with 'ghcr.io/azure/kdebug:main' to the target pod."`
	Namespace string `long:"namespace" description:"the namespace of the pod."`
	Command   string `long:"command" description:"Customize the command to be run in container namespace. Leave it blank to use 'sh'."`
	Image     string `long:"image" description:"Customize the image to be used to run command when using --pod. Leave it blank to use busybox."`
}

type NetexecTool struct {
	pid       string
	podName   string
	namespace string
	command   string
	image     string
}

const (
	DefaultCommand                   = "sh"
	DefaultContainerImage            = "busybox"
	DefaultNamespace                 = "default"
	DefaultKubectlBasicCommandFormat = "debug -ti %s --image %s -n %s -- "
)

func New() *NetexecTool {
	return &NetexecTool{}
}

func (c *NetexecTool) Name() string {
	return "Netexec"
}

func logAndExec(name string, args ...string) *exec.Cmd {
	log.Infof("Exec %s %+v", name, args)
	return exec.Command(name, args...)
}

func (c *NetexecTool) Run(ctx *base.ToolContext) error {
	if err := c.parseAndCheckParameters(ctx); err != nil {
		return err
	}

	if len(c.pid) > 0 {
		err := c.checkWithPid()
		if err != nil {
			return err
		}

		return nil
	}

	err := c.checkWithPod(ctx.KubeConfigFlag)
	if err != nil {
		return err
	}

	return nil
}

func (c *NetexecTool) ParseArgs(ctx *base.ToolContext, args []string) error {
	var config Config
	remainingArgs, err := flags.ParseArgs(&config, args)
	if err != nil {
		return err
	}
	ctx.Config = &config
	ctx.Args = remainingArgs
	return nil
}

func (c *NetexecTool) parseAndCheckParameters(ctx *base.ToolContext) error {
	config := ctx.Config.(*Config)

	if len(config.Pid) == 0 && len(config.PodName) == 0 {
		return fmt.Errorf("Either --pid and --pod should be set.")
	}
	if len(config.Pid) > 0 && len(config.PodName) > 0 {
		return fmt.Errorf("--pid and --pod can not be assigned together. Please set either of them.")
	}
	if len(config.PodName) > 0 {
		if ctx.KubeConfigFlag == nil {
			return fmt.Errorf("kubernetes client is not availble. Check kubeconfig.")
		}
	}

	c.pid = config.Pid
	c.podName = config.PodName
	if len(config.Command) > 0 {
		c.command = config.Command
	} else {
		c.command = DefaultCommand
	}

	if len(config.Image) > 0 {
		c.image = config.Image
	} else {
		c.image = DefaultContainerImage
	}

	if len(config.Namespace) > 0 {
		c.namespace = config.Namespace
	} else {
		c.namespace = DefaultNamespace
	}

	return nil
}

func (c *NetexecTool) checkWithPid() error {
	_, err := logAndExec("nsenter", "-n", "-t", c.pid).Output()
	if err != nil {
		return err
	}

	args := strings.Fields(c.command)
	cmd := logAndExec(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (c *NetexecTool) checkWithPod(configFlags *genericclioptions.ConfigFlags) error {
	cmd := fmt.Sprintf("%s%s", fmt.Sprintf(DefaultKubectlBasicCommandFormat, c.podName, c.image, c.namespace), c.command)
	arg := strings.Fields(cmd)
	log.Infof("The command is equivalent to 'kubectl %s'", cmd)
	kubectlCmd := kubecmd.NewKubectlCommand(kubecmd.KubectlOptions{
		ConfigFlags: configFlags,
		IOStreams:   genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	})
	kubectlCmd.SetArgs(arg)

	err := kubectlCmd.Execute()
	if err != nil {
		return err
	}

	return nil
}
