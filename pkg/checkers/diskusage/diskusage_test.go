package diskusage

import (
	"testing"
)

func TestDfParse_Success(t *testing.T) {
	dfOutput := `Filesystem      Size  Used Avail Use% Mounted on
	/dev/sdb        251G   11G  228G   5% /`

	result, _ := parseDfResult(dfOutput, DfHeaders["LINUX"])
	if len(result) != 1 {
		t.Errorf("Expect the length of result is 1 but got %+v", len(result))
	}

	if result[0].Filesystem != "/dev/sdb" {
		t.Errorf("Expect Filesystem is /dev/sdb but got %s", result[0].Filesystem)
	}

	if result[0].Size != "251G" {
		t.Errorf("Expect Size is 251G but got %s", result[0].Size)
	}

	if result[0].Used != "11G" {
		t.Errorf("Expect Used is 11G but got %s", result[0].Used)
	}

	if result[0].Avail != "228G" {
		t.Errorf("Expect Avail is 228G but got %s", result[0].Avail)
	}

	if result[0].Use != 5 {
		t.Errorf("Expect Use is 5 but got %v", result[0].Use)
	}

	if result[0].MountedOn != "/" {
		t.Errorf("Expect MountedOn is / but got %s", result[0].MountedOn)
	}
}

func TestDfParse_FreeBSD_Success(t *testing.T) {
	dfOutput := `Filesystem         Size    Used   Avail Capacity  Mounted on
	/dev/gpt/rootfs     29G    4.0G     23G    15%    /`

	result, _ := parseDfResult(dfOutput, DfHeaders["FREEBSD"])
	if len(result) != 1 {
		t.Errorf("Expect the length of result is 1 but got %+v", len(result))
	}

	if result[0].Filesystem != "/dev/gpt/rootfs" {
		t.Errorf("Expect Filesystem is /dev/gpt/rootfs but got %s", result[0].Filesystem)
	}

	if result[0].Size != "29G" {
		t.Errorf("Expect Size is 29G but got %s", result[0].Size)
	}

	if result[0].Used != "4.0G" {
		t.Errorf("Expect Used is 4.0G but got %s", result[0].Used)
	}

	if result[0].Avail != "23G" {
		t.Errorf("Expect Avail is 23G but got %s", result[0].Avail)
	}

	if result[0].Use != 15 {
		t.Errorf("Expect Use is 15 but got %v", result[0].Use)
	}

	if result[0].MountedOn != "/" {
		t.Errorf("Expect MountedOn is / but got %s", result[0].MountedOn)
	}
}

func TestDfParse_Failed(t *testing.T) {
	dfOutput := `Filesystem      Size  Used Avail Use% MountedOn
	/dev/sdb        251G   11G  228G   5% /`

	_, err := parseDfResult(dfOutput, DfHeaders["LINUX"])
	if err == nil {
		t.Errorf("Expect error in parseDfResult but not")
	}
}
