package systemload

import (
	"testing"
)

func TestMemPercentage_Success(t *testing.T) {
	usage := getMemPercentage(30, 100)
	if usage != 70 {
		t.Errorf("Expect the mem percentage is 70 but got %f", usage)
	}
}

func TestSystemCPUPercentage_Success(t *testing.T) {
	usage := getSystemCPUPercentage(2000, 5000)
	if usage != 60 {
		t.Errorf("Expect the cpu percentage is 60 but got %f", usage)
	}
}

func TestProcessCPUPercentageAsGlobal_Success(t *testing.T) {
	usage := getProcessCPUPercentageAsGlobal(50, 5000)
	if usage != 1 {
		t.Errorf("Expect the process cpu percentage is 1 but got %f", usage)
	}
}

func TestProcessCPUPercentageAsSingleCore_Success(t *testing.T) {
	usage := getProcessCPUPercentageAsSingleCore(400, 2)
	if usage != 200 {
		t.Errorf("Expect the process cpu percentage is 200 but got %f", usage)
	}

	usage = getProcessCPUPercentageAsSingleCore(100, 10)
	if usage != 10 {
		t.Errorf("Expect the process cpu percentage is 10 but got %f", usage)
	}
}
