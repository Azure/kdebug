package oom

import (
	"fmt"
	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
	"io/ioutil"
	"os"
	"testing"
)

const testString = "Feb 22 16:15:02 k8s-ingress-11186066-z1-vmss0000B3 kernel: [989751.247878] Memory cgroup out of memory: Killed process 3841 (nginx) total-vm:240652kB, anon-rss:130344kB, file-rss:5212kB, shmem-rss:208kB, UID:101 pgtables:332kB oom_score_adj:986\n"

func TestCheckOOMLogWhenOOM(t *testing.T) {
	environment := &env.StaticEnvironment{
		Flags: []string{"ubuntu"},
	}
	if !envCheck(env.GetEnvironment()) {
		fmt.Println("skip oom test")
		return
	}
	tmp, err := ioutil.TempFile("", "kernlog")
	if err != nil {
		t.Fatalf("error creating tmp file:%v", err)
	}
	check := OOMChecker{kernLogPath: tmp.Name()}
	defer func() {
		e := os.Remove(check.kernLogPath)
		if e != nil {
			t.Errorf(e.Error())
		}
	}()
	//should be 600. But it fails in 600
	err = os.WriteFile(check.kernLogPath, []byte(testString), 777)
	if err != nil {
		t.Errorf("Create tmp file error:%v", err)
	}
	result, _ := check.Check(&base.CheckContext{
		Environment: environment,
	})
	if len(result) != 1 {
		t.Errorf("Get unexpected OOM result length %v", len(result))
	}
	checkErr := result[0].Error
	if checkErr != "progress:[3841 nginx] is OOM kill at time [Feb 22 16:15:02]. [rss:130344kB] [oom_score_adj:986]\n" {
		t.Errorf("Unexpected check result:\n %v \n %v", result[0].Description, checkErr)
	}

}
