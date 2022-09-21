package netexec

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
	log "github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kubecmd "k8s.io/kubectl/pkg/cmd"
)

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

func (c *NetexecTool) parseAndCheckParameters(ctx *base.ToolContext) error {
	if len(ctx.Netexec.Pid) == 0 && len(ctx.Netexec.PodName) == 0 {
		return fmt.Errorf("Either --netexec.pid and --netexec.pod should be set.")
	}
	if len(ctx.Netexec.Pid) > 0 && len(ctx.Netexec.PodName) > 0 {
		return fmt.Errorf("--netexec.pid and --netexec.pod can not be assigned together. Please set either of them.")
	}
	if len(ctx.Netexec.PodName) > 0 {
		if ctx.KubeConfigFlag == nil {
			return fmt.Errorf("kubernetes client is not availble. Check kubeconfig.")
		}
	}

	c.pid = ctx.Netexec.Pid
	c.podName = ctx.Netexec.PodName
	if len(ctx.Netexec.Command) > 0 {
		c.command = ctx.Netexec.Command
	} else {
		c.command = DefaultCommand
	}

	if len(ctx.Netexec.Image) > 0 {
		c.image = ctx.Netexec.Image
	} else {
		c.image = DefaultContainerImage
	}

	if len(ctx.Netexec.Namespace) > 0 {
		c.namespace = ctx.Netexec.Namespace
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
