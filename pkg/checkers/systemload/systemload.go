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
	ProcessCPUTooHigh            = "The CPU usage of process [%d] (%s) is higher than limit. The proportion of cpu is %f%% to whole capacity (limit is %f%%). The proportion of cpu is %f%% to one core (limit is %f%%)"
	GloablHighCPURecommandation  = "You may remote to the target VM and use 'top' to find out which process consumes most of CPU."
	ProcessHighCPURecommandation = "You may restart to process if feasible and see whether the CPU usage comes to normal. Or you can 'perf' to diagnose the root cause."
)

var (
	GlobalCPUPercentageLimit float64 = 80  // The percentage compare to the whole VM CPU capacity. 100 means using up all the cpu capacity
	ClkTck                   float64 = 100 // default
	InterestedProcNames              = map[string]ProcLimitMeasurement{
		"etcd":    {CPULimitAsGloabl: 50, CPULimitAsSingleCore: 80},
		"kubelet": {CPULimitAsGloabl: 50, CPULimitAsSingleCore: 80}}
)

type InterestedProc struct {
	StatFilePath         string  // Process stat file location. Should follow /proc/[pid]/stat
	Name                 string  // The command of the process
	Pid                  uint64  // Pid
	TotalTime            uint64  // Time of the process used in cpu cycle
	CPULimitAsGloabl     float64 // CPU limit compare to the whole VM CPU capacity
	CPULimitAsSingleCore float64 // CPU limit compare to one core
}

type ProcLimitMeasurement struct {
	CPULimitAsGloabl     float64 // The percentage compare to the whole VM CPU capacity. 100 means using up all the cpu capacity
	CPULimitAsSingleCore float64 // The percentage compare to one core. 100 means using up 1 core's capacity. Maximum number can be 100 * cores
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

	procStatusFiles, err := filepath.Glob("/proc/[0-9]*/stat")
	interestedProcesses := []*InterestedProc{}

	// Read status and find out interested process
	for _, f := range procStatusFiles {
		stat, err := linuxproc.ReadProcessStat(f)
		if err != nil {
			continue
		}

		var cmd = stat.Comm[1 : len(stat.Comm)-1] // name: (cmd)
		if limit, ok := InterestedProcNames[cmd]; ok {
			interestedProcesses = append(interestedProcesses, &InterestedProc{
				StatFilePath:         f,
				Name:                 cmd,
				Pid:                  stat.Pid,
				CPULimitAsGloabl:     limit.CPULimitAsGloabl,
				CPULimitAsSingleCore: limit.CPULimitAsSingleCore,
				TotalTime:            stat.Utime + stat.Stime, // Time in user space + Time in kernal space
			})
		}
	}

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

	if usage > GlobalCPUPercentageLimit {
		result = append(result, &base.CheckResult{
			Checker:     c.Name(),
			Error:       fmt.Sprintf(GlobalCPUTooHigh, GlobalCPUPercentageLimit, usage),
			Description: GloablHighCPURecommandation,
		})
	}

	// Calculate interested proc
	for _, proc := range interestedProcesses {
		stat, err := linuxproc.ReadProcessStat(proc.StatFilePath)
		if err != nil {
			continue
		}

		// https://stackoverflow.com/questions/16726779/how-do-i-get-the-total-cpu-usage-of-an-application-from-proc-pid-stat/16736599#16736599
		totalTime := stat.Utime + stat.Stime
		cpuUsageAsGlobal := 100 * float64(totalTime-proc.TotalTime) / float64(deltaTotalTime)
		cpuUsageAsSingleCore := 100 * float64(totalTime-proc.TotalTime) / ClkTck

		if cpuUsageAsGlobal > proc.CPULimitAsGloabl || cpuUsageAsSingleCore > proc.CPULimitAsSingleCore {
			result = append(result, &base.CheckResult{
				Checker:     c.Name(),
				Error:       fmt.Sprintf(ProcessCPUTooHigh, proc.Pid, proc.Name, cpuUsageAsGlobal, proc.CPULimitAsGloabl, cpuUsageAsSingleCore, proc.CPULimitAsSingleCore),
				Description: ProcessHighCPURecommandation,
			})
		}
	}

	return result, nil
}

func GetTotalTime(stat linuxproc.CPUStat) uint64 {
	return stat.User + stat.Nice + stat.System + stat.Idle + stat.IOWait + stat.IRQ + stat.SoftIRQ +
		stat.Steal + stat.Guest + stat.GuestNice
}
