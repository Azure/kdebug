package pod

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
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

	name, namespace := ctx.Pod.Name, ctx.Pod.Namespace
	pod, err := ctx.KubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		eventsInterface := ctx.KubeClient.CoreV1().Events(namespace)
		selector := eventsInterface.GetFieldSelector(&name, &namespace, nil, nil)
		initialOpts := metav1.ListOptions{
			FieldSelector: selector.String(),
			Limit:         util.DefaultChunkSize,
		}
		events := &corev1.EventList{}
		err2 := runtimeresource.FollowContinue(&initialOpts,
			func(options metav1.ListOptions) (runtime.Object, error) {
				newList, err := eventsInterface.List(context.TODO(), options)
				if err != nil {
					return nil, runtimeresource.EnhanceListError(err, options, "events")
				}
				events.Items = append(events.Items, newList.Items...)
				return newList, nil
			})

		if err2 != nil {
			return []*base.CheckResult{
				{
					Checker: c.Name(),
					Error:   err.Error(),
				},
			}, nil
		}

		if len(events.Items) == 0 {
			return []*base.CheckResult{
				{
					Checker:     c.Name(),
					Description: fmt.Sprintf("Nothing found for %s/%s", namespace, name),
				},
			}, nil
		}

		text, err3 := tabbedString(func(out io.Writer) error {
			w := describe.NewPrefixWriter(out)
			w.Write(describe.LEVEL_0, "Pod '%v': error '%v', but found events.\n", name, err)
			describe.DescribeEvents(events, w)
			return nil
		})
		if err3 != nil {
			return nil, err3
		}

		return []*base.CheckResult{
			{
				Checker:     c.Name(),
				Description: fmt.Sprintf("Information for %s/%s:\n%s", namespace, name, text),
			},
		}, nil
	}

	var events *corev1.EventList
	ref, err := reference.GetReference(scheme.Scheme, pod)
	if err != nil {
		log.Errorf("Unable to construct reference to '%#v': %v", pod, err)
		return []*base.CheckResult{
			{
				Checker: c.Name(),
				Error:   err.Error(),
			},
		}, nil
	}

	ref.Kind = ""
	if _, isMirrorPod := pod.Annotations[corev1.MirrorPodAnnotationKey]; isMirrorPod {
		ref.UID = types.UID(pod.Annotations[corev1.MirrorPodAnnotationKey])
	}
	events, _ = searchEvents(ctx.KubeClient.CoreV1(), ref, util.DefaultChunkSize)
	text, err := describePodStatus(pod, events)
	return []*base.CheckResult{
		{
			Checker:     c.Name(),
			Description: fmt.Sprintf("Information for %s/%s:\n%s", namespace, name, text),
		},
	}, nil
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

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
