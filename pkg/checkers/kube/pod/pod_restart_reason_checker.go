package pod

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/duration"
	runtimeresource "k8s.io/cli-runtime/pkg/resource"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/describe"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/qos"
)

const levelSpace = "  "

type KubePodRestartReasonChecker struct {
}

func New() *KubePodRestartReasonChecker {
	return &KubePodRestartReasonChecker{}
}

func (c *KubePodRestartReasonChecker) Name() string {
	return "KubePodRestartReason"
}

// Check borrows many logic and helper functions from src/k8s.io/kubectl/pkg/describe to check Pod status and events.
func (c *KubePodRestartReasonChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	if ctx.KubeClient == nil {
		log.Warn("Skip KubePodRestartReasonChecker due to missing kube client")
		return nil, nil
	}

	pods, err := ctx.KubeClient.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Fail to list pods")
		return nil, err
	}

	results := []*base.CheckResult{}
	for _, pod := range pods.Items {
		var crashing = false
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
				crashing = true
				break
			}
		}

		if crashing {
			result := c.checkPod(ctx, &pod)
			if result != nil {
				results = append(results, result)
			}
		}
	}

	return results, nil
}

func (c *KubePodRestartReasonChecker) checkPod(ctx *base.CheckContext, pod *v1.Pod) *base.CheckResult {
	var events *corev1.EventList
	ref, err := reference.GetReference(scheme.Scheme, pod)
	if err != nil {
		log.WithFields(log.Fields{"pod": pod, "error": err}).Warn("Unable to construct reference")
		return nil
	}

	ref.Kind = ""
	if _, isMirrorPod := pod.Annotations[corev1.MirrorPodAnnotationKey]; isMirrorPod {
		ref.UID = types.UID(pod.Annotations[corev1.MirrorPodAnnotationKey])
	}
	events, _ = searchEvents(ctx.KubeClient.CoreV1(), ref, util.DefaultChunkSize)
	text, _ := describePodStatus(pod, events)
	logs := strings.Split(text, "\n")

	for i := range logs {
		logs[i] = levelSpace + logs[i]
	}

	return &base.CheckResult{
		Checker:     c.Name(),
		Error:       fmt.Sprintf("one or more containers of %s/%s are failing and restarting repeatedly.", pod.Namespace, pod.Name),
		Description: fmt.Sprintf("%s/%s is not running well.", pod.Namespace, pod.Name),
		Logs:        logs,
	}
}

func describePodStatus(pod *corev1.Pod, events *corev1.EventList) (string, error) {
	return tabbedString(func(out io.Writer) error {
		w := describe.NewPrefixWriter(out)
		w.Write(describe.LEVEL_0, "Name:\t%s\n", pod.Name)
		w.Write(describe.LEVEL_0, "Namespace:\t%s\n", pod.Namespace)
		if pod.Status.StartTime != nil {
			w.Write(describe.LEVEL_0, "Start Time:\t%s\n", pod.Status.StartTime.Time.Format(time.RFC1123Z))
		}
		if pod.DeletionTimestamp != nil {
			w.Write(describe.LEVEL_0, "Status:\tTerminating (lasts %s)\n", translateTimestampSince(*pod.DeletionTimestamp))
			w.Write(describe.LEVEL_0, "Termination Grace Period:\t%ds\n", *pod.DeletionGracePeriodSeconds)
		} else {
			w.Write(describe.LEVEL_0, "Status:\t%s\n", string(pod.Status.Phase))
		}
		if len(pod.Status.Reason) > 0 {
			w.Write(describe.LEVEL_0, "Reason:\t%s\n", pod.Status.Reason)
		}
		if len(pod.Status.Message) > 0 {
			w.Write(describe.LEVEL_0, "Message:\t%s\n", pod.Status.Message)
		}
		describeContainers("Containers", pod.Spec.Containers, pod.Status.ContainerStatuses, describe.EnvValueRetriever(pod), w, "")
		if len(pod.Status.Conditions) > 0 {
			w.Write(describe.LEVEL_0, "Conditions:\n  Type\tStatus\n")
			for _, c := range pod.Status.Conditions {
				w.Write(describe.LEVEL_1, "%v \t%v \n",
					c.Type,
					c.Status)
			}
		}
		if pod.Status.QOSClass != "" {
			w.Write(describe.LEVEL_0, "QoS Class:\t%s\n", pod.Status.QOSClass)
		} else {
			w.Write(describe.LEVEL_0, "QoS Class:\t%s\n", qos.GetPodQOS(pod))
		}
		if events != nil {
			describe.DescribeEvents(events, w)
		}
		return nil
	})
}

func tabbedString(f func(io.Writer) error) (string, error) {
	out := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	out.Init(buf, 0, 8, 2, ' ', 0)

	err := f(out)
	if err != nil {
		return "", err
	}

	out.Flush()
	str := string(buf.String())
	return str, nil
}

func searchEvents(client corev1client.EventsGetter, objOrRef runtime.Object, limit int64) (*corev1.EventList, error) {
	ref, err := reference.GetReference(scheme.Scheme, objOrRef)
	if err != nil {
		return nil, err
	}
	stringRefKind := string(ref.Kind)
	var refKind *string
	if len(stringRefKind) > 0 {
		refKind = &stringRefKind
	}
	stringRefUID := string(ref.UID)
	var refUID *string
	if len(stringRefUID) > 0 {
		refUID = &stringRefUID
	}

	e := client.Events(ref.Namespace)
	fieldSelector := e.GetFieldSelector(&ref.Name, &ref.Namespace, refKind, refUID)
	initialOpts := metav1.ListOptions{FieldSelector: fieldSelector.String(), Limit: limit}
	eventList := &corev1.EventList{}
	err = runtimeresource.FollowContinue(&initialOpts,
		func(options metav1.ListOptions) (runtime.Object, error) {
			newEvents, err := e.List(context.TODO(), options)
			if err != nil {
				return nil, runtimeresource.EnhanceListError(err, options, "events")
			}
			eventList.Items = append(eventList.Items, newEvents.Items...)
			return newEvents, nil
		})
	return eventList, err
}

func describeContainers(label string, containers []corev1.Container, containerStatuses []corev1.ContainerStatus,
	resolverFn describe.EnvVarResolverFunc, w describe.PrefixWriter, space string) {
	statuses := map[string]corev1.ContainerStatus{}
	for _, status := range containerStatuses {
		statuses[status.Name] = status
	}

	for _, container := range containers {
		status, ok := statuses[container.Name]
		describeContainerBasicInfo(container, status, ok, space, w)
		if ok {
			describeContainerState(status, w)
		}
	}
}

func describeContainerBasicInfo(container corev1.Container, status corev1.ContainerStatus, ok bool, space string, w describe.PrefixWriter) {
	nameIndent := ""
	if len(space) > 0 {
		nameIndent = " "
	}
	w.Write(describe.LEVEL_1, "%s%v:\n", nameIndent, container.Name)
	if ok {
		w.Write(describe.LEVEL_2, "Container ID:\t%s\n", status.ContainerID)
	}
	w.Write(describe.LEVEL_2, "Image:\t%s\n", container.Image)
	if ok {
		w.Write(describe.LEVEL_2, "Image ID:\t%s\n", status.ImageID)
	}
}

func describeContainerState(status corev1.ContainerStatus, w describe.PrefixWriter) {
	describeStatus("State", status.State, w)
	if status.LastTerminationState.Terminated != nil {
		describeStatus("Last State", status.LastTerminationState, w)
	}
	w.Write(describe.LEVEL_2, "Ready:\t%v\n", printBool(status.Ready))
	w.Write(describe.LEVEL_2, "Restart Count:\t%d\n", status.RestartCount)
}

func describeStatus(stateName string, state corev1.ContainerState, w describe.PrefixWriter) {
	switch {
	case state.Running != nil:
		w.Write(describe.LEVEL_2, "%s:\tRunning\n", stateName)
		w.Write(describe.LEVEL_3, "Started:\t%v\n", state.Running.StartedAt.Time.Format(time.RFC1123Z))
	case state.Waiting != nil:
		w.Write(describe.LEVEL_2, "%s:\tWaiting\n", stateName)
		if state.Waiting.Reason != "" {
			w.Write(describe.LEVEL_3, "Reason:\t%s\n", state.Waiting.Reason)
		}
	case state.Terminated != nil:
		w.Write(describe.LEVEL_2, "%s:\tTerminated\n", stateName)
		if state.Terminated.Reason != "" {
			w.Write(describe.LEVEL_3, "Reason:\t%s\n", state.Terminated.Reason)
		}
		if state.Terminated.Message != "" {
			w.Write(describe.LEVEL_3, "Message:\t%s\n", state.Terminated.Message)
		}
		w.Write(describe.LEVEL_3, "Exit Code:\t%d\n", state.Terminated.ExitCode)
		if state.Terminated.Signal > 0 {
			w.Write(describe.LEVEL_3, "Signal:\t%d\n", state.Terminated.Signal)
		}
		w.Write(describe.LEVEL_3, "Started:\t%s\n", state.Terminated.StartedAt.Time.Format(time.RFC1123Z))
		w.Write(describe.LEVEL_3, "Finished:\t%s\n", state.Terminated.FinishedAt.Time.Format(time.RFC1123Z))
	default:
		w.Write(describe.LEVEL_2, "%s:\tWaiting\n", stateName)
	}
}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

func printBool(value bool) string {
	if value {
		return "True"
	}

	return "False"
}
