package systemload

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	linuxproc "github.com/c9s/goprocinfo/linux"
)

const (
	GlobalCPUTooHigh             = "The VM's CPU usage is higher then limit %f. It's currently %f."
	HighUsageRecommandation      = "Check files listed. If it's just log files or can be deleted, run bash command: `truncate -s 0 /path/to/file` to reduce disk usage. Note: `rm` will not really delete the file if it's opened by processes."
	FailedToCheckDiskUsageWithDf = "Failed to check disk usage with 'df -h'"
)

var (
	GlobalCPUPercentageLimit            = float64(80)
	GlobalMemoryPercentageLimit         = 80
	ClkTck                      float64 = 100 // default
	InsterestedProcNames                = map[string]bool{"etcd": true, "kubelet": true}
)

type InterestedProc struct {
	StatFilePath  string
	Name          string
	Pid           uint64
	TotalTime     uint64
	ProcessState  linuxproc.ProcessStat
	Uptime        linuxproc.Uptime
	CPUAlertLimit float64
}

type SystemLoadChecker struct {
}

func New() *SystemLoadChecker {
	return &SystemLoadChecker{}
}

func (c *SystemLoadChecker) Name() string {
	return "SystemLoad"
}

func (c *SystemLoadChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	result := []*base.CheckResult{}

	procStatusFiles, err := filepath.Glob("/proc/[0-9]+/stat")
	interestedProcesses := []*InterestedProc{}

	// Read status and find out interested process
	for _, f := range procStatusFiles {
		stat, err := linuxproc.ReadProcessStat(f)
		if err != nil {
			continue
		}

		if InsterestedProcNames[stat.Comm] {
			interestedProcesses = append(interestedProcesses, &InterestedProc{
				StatFilePath: f,
				Name:         stat.Comm,
				Pid:          stat.Pid,
				TotalTime:    stat.Utime + stat.Stime, // Time in user space + Time in kernal space
			})
		}
	}

	// https://stackoverflow.com/questions/16726779/how-do-i-get-the-total-cpu-usage-of-an-application-from-proc-pid-stat/16736599#16736599
	// Read global status
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		return result, err
	}

	// How to calculate global cpu usage: https://rosettacode.org/wiki/Linux_CPU_utilization
	var previousIdleTime = stat.CPUStatAll.Idle
	var previousTotalTime = GetTotalTime(stat.CPUStatAll)

	time.Sleep(time.Second)

	stat, err = linuxproc.ReadStat("/proc/stat")
	if err != nil {
		return result, err
	}

	var deltaIdleTime = stat.CPUStatAll.Idle - previousIdleTime
	var deltaTotalTime = GetTotalTime(stat.CPUStatAll) - previousTotalTime
	var usage = 100 - (float64(100*(deltaIdleTime)) / float64(deltaTotalTime))
	fmt.Print(usage)

	// Calculate interested proc
	for _, proc := range interestedProcesses {
		stat, err := linuxproc.ReadProcessStat(proc.StatFilePath)
		if err != nil {
			continue
		}

		totalTime := stat.Utime + stat.Stime
		usage := float64(totalTime-proc.TotalTime) / float64(deltaTotalTime)

		print(fmt.Sprintf("%s usage: %f", proc.Name, usage))
	}

	if usage > GlobalCPUPercentageLimit {
		result = append(result, &base.CheckResult{
			Checker: c.Name(),
			Error:   fmt.Sprintf(GlobalCPUTooHigh, GlobalCPUPercentageLimit, usage),
		})
	}

	return result, nil
}

func GetTotalTime(stat linuxproc.CPUStat) uint64 {
	return stat.User + stat.Nice + stat.System + stat.Idle + stat.IOWait + stat.IRQ + stat.SoftIRQ +
		stat.Steal + stat.Guest + stat.GuestNice
}
