package dns

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestCheckObjectSize_OK(t *testing.T) {
	cm := v1.ConfigMap{
		BinaryData: map[string][]byte{
			"key": make([]byte, 100),
		},
	}
	checker := New()
	result := checker.checkObjectSize("ConfigMap", "default", "cm", cm)
	if !result.Ok() {
		t.Errorf("Expect ok result but got %+v", result)
	}
}

func TestCheckObjectSize_Warn(t *testing.T) {
	cm := v1.ConfigMap{
		BinaryData: map[string][]byte{
			"key": make([]byte, WarnSizeThreshold+1),
		},
	}
	checker := New()
	result := checker.checkObjectSize("ConfigMap", "default", "cm", cm)
	if result.Ok() {
		t.Errorf("Expect non ok result but got %+v", result)
	}
	if result.Error == "" || result.Description == "" || len(result.Recommendations) == 0 {
		t.Errorf("Expect non empty result but got %+v", result)
	}
}
