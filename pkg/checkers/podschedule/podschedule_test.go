package podschedule

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestPodSchedule_Single_Panic(t *testing.T) {
	podList := []v1.Pod{
		{
			Spec: v1.PodSpec{
				NodeName: "a",
			},
		},
	}

	defer func() {
		if recover() == nil {
			t.Errorf("Expect panic")
		}
	}()

	checker := New()
	checker.checkPodsScheduleInReplicaSet("rc1", podList)
}

func TestPodSchedule_DifferentName_OK(t *testing.T) {
	podList := []v1.Pod{
		{
			Spec: v1.PodSpec{
				NodeName: "a",
			},
		},
		{
			Spec: v1.PodSpec{
				NodeName: "b",
			},
		},
	}

	checker := New()
	result := checker.checkPodsScheduleInReplicaSet("rc1", podList)
	if !result.Ok() {
		t.Errorf("Expect ok result but got %+v", result)
	}
}

func TestPodSchedule_Failed(t *testing.T) {
	podList := []v1.Pod{
		{
			Spec: v1.PodSpec{
				NodeName: "a",
			},
		},
		{
			Spec: v1.PodSpec{
				NodeName: "a",
			},
		},
	}

	checker := New()
	result := checker.checkPodsScheduleInReplicaSet("rc1", podList)
	if result.Ok() {
		t.Errorf("Expect failed result but got %+v", result)
	}
}
