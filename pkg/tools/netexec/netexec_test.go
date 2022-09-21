package netexec

import (
	"testing"

	"github.com/Azure/kdebug/pkg/base"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestParseParameterPid_Success(t *testing.T) {
	netexec := &NetexecTool{}
	netexec.parseAndCheckParameters(&base.ToolContext{
		Config: &Config{
			Pid:     "1",
			Command: "bash",
		},
	})

	if netexec.pid != "1" {
		t.Errorf("pid should got %s but got %s", "1", netexec.pid)
	}

	if netexec.command != "bash" {
		t.Errorf("command should got %s but got %s", "bash", netexec.command)
	}
}

func TestParseParameterPod_Success(t *testing.T) {
	netexec := &NetexecTool{}
	netexec.parseAndCheckParameters(&base.ToolContext{
		Config: &Config{
			PodName:   "pod",
			Command:   "bash",
			Namespace: "kube-system",
			Image:     "image",
		},
		KubeConfigFlag: &genericclioptions.ConfigFlags{},
	})

	if netexec.podName != "pod" {
		t.Errorf("podname should got %s but got %s", "pod", netexec.podName)
	}

	if netexec.command != "bash" {
		t.Errorf("command should got %s but got %s", "bash", netexec.command)
	}

	if netexec.namespace != "kube-system" {
		t.Errorf("namespace should got %s but got %s", "kube-system", netexec.namespace)
	}

	if netexec.image != "image" {
		t.Errorf("image should got %s but got %s", "image", netexec.image)
	}
}

func TestParseParameter_Failed(t *testing.T) {
	netexec := &NetexecTool{}
	err := netexec.parseAndCheckParameters(&base.ToolContext{
		Config: &Config{},
	})

	if err == nil {
		t.Error("Should got err: 'Either --pid and --pod should be set.', but error is not raised")
	}

	err = netexec.parseAndCheckParameters(&base.ToolContext{
		Config: &Config{
			Pid:     "1",
			PodName: "pod",
		},
	})

	if err == nil {
		t.Error("Should got err: '--pid and --pod can not be assigned together. Please set either of them.', but error is not raised")
	}

	err = netexec.parseAndCheckParameters(&base.ToolContext{
		Config: &Config{
			PodName:   "pod",
			Command:   "bash",
			Namespace: "kube-system",
			Image:     "image",
		},
	})

	if err == nil {
		t.Error("Should got err: 'kubernetes client is not availble. Check kubeconfig.', but error is not raised")
	}
}
