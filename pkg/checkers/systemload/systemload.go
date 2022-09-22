package systemload

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	linuxproc "github.com/c9s/goprocinfo/linux"
)

const (
	GlobalCPUTooHigh               = "The VM's CPU usage is higher than threshold. Currently %.1f%% (threshold is %.1f%%)."
	GlobalMemoryTooHigh            = "The VM's Memory usage is higher than threshold. Currently %.1f%% (threshold is %.1f%%)"
	ProcessCPUTooHigh              = "The CPU usage of process [%d] (%s) is higher than threshold. The proportion of cpu is %.1f%% to whole capacity (threshold is %.1f%%). The proportion of cpu is %.1f%% to one core (threshold is %.1f%%)"
	GloablHighCPURecommandation    = "You may remote to the target VM and use 'top' to find out which process consumes most of CPU. Further actions may depends."
	GloablHighMemoryRecommandation = "You may remote to the target VM and use 'top' to find out which process consumes most of Memory. Further actions may depends."
	ProcessHighCPURecommandation   = "You may restart to process if feasible and see whether the CPU usage comes to normal. Or you can 'perf' to diagnose the root cause."
)

var (
	VMCPUPercentageLimit    float64 = 80  // The percentage compare to the whole VM CPU capacity. 100 means using up all the cpu capacity
	VMMemoryPercentageLimit float64 = 90  // The percentage compare to the VM Total Memory. 100 means using up all the memory capacity
	ClkTck                  float64 = 100 // default value of cycles per seconds
	CPUSpan                 float64 = 1   // The timespan of CPU load in seconds
	InterestedProcNames             = map[string]ProcLimitMeasurement{
		"etcd":           {CPULimitAsGloabl: 50, CPULimitAsSingleCore: 80},
		"kubelet":        {CPULimitAsGloabl: 50, CPULimitAsSingleCore: 80},
		"kube-apiserver": {CPULimitAsGloabl: 50, CPULimitAsSingleCore: 80}}
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

	if !ctx.Environment.HasFlag("linux") {
		return result, nil
	}

	// VM Memory
	memInfo, err := linuxproc.ReadMemInfo("/proc/meminfo")
	if err != nil {
		return result, err
	}
	var memUsage = getMemPercentage(memInfo.MemAvailable, memInfo.MemTotal)
	if memUsage > VMMemoryPercentageLimit {
		result = append(result, &base.CheckResult{
			Checker:     c.Name(),
			Error:       fmt.Sprintf(GlobalMemoryTooHigh, memUsage, VMMemoryPercentageLimit),
			Description: GloablHighMemoryRecommandation,
		})
	}

	interestedProcesses, err := getInterestedProc()
	if err != nil {
		return result, err
	}

	// Read global status
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		return result, err
	}

	// How to calculate global cpu usage: https://rosettacode.org/wiki/Linux_CPU_utilization
	previousIdleTime, previousTotalTime := getSystemCPUTime(stat.CPUStatAll)

	// Sleep a time span and check cpu time again to get average CPU load
	time.Sleep(time.Duration(CPUSpan * float64(time.Second)))

	stat, err = linuxproc.ReadStat("/proc/stat")
	if err != nil {
		return result, err
	}

	idleTime, totalTime := getSystemCPUTime(stat.CPUStatAll)
	var deltaSystemIdleTime = idleTime - previousIdleTime
	var deltaSystemTotalTime = totalTime - previousTotalTime
	var usage = getSystemCPUPercentage(deltaSystemIdleTime, deltaSystemTotalTime)

	// VM CPU
	if usage > VMCPUPercentageLimit {
		result = append(result, &base.CheckResult{
			Checker:     c.Name(),
			Error:       fmt.Sprintf(GlobalCPUTooHigh, usage, VMCPUPercentageLimit),
			Description: GloablHighCPURecommandation,
		})
	}

	// Interested proc cpu
	for _, proc := range interestedProcesses {
		stat, err := linuxproc.ReadProcessStat(proc.StatFilePath)
		if err != nil {
			continue
		}

		// https://stackoverflow.com/questions/16726779/how-do-i-get-the-total-cpu-usage-of-an-application-from-proc-pid-stat/16736599#16736599
		totalTime := stat.Utime + stat.Stime
		cpuUsageAsGlobal := getProcessCPUPercentageAsGlobal(totalTime-proc.TotalTime, deltaSystemTotalTime)
		cpuUsageAsSingleCore := getProcessCPUPercentageAsSingleCore(totalTime-proc.TotalTime, CPUSpan)

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

func getTotalTime(stat linuxproc.CPUStat) uint64 {
	return stat.User + stat.Nice + stat.System + stat.Idle + stat.IOWait + stat.IRQ + stat.SoftIRQ +
		stat.Steal + stat.Guest + stat.GuestNice
}

func getInterestedProc() ([]*InterestedProc, error) {
	result := []*InterestedProc{}

	procStatusFiles, err := filepath.Glob("/proc/[0-9]*/stat")
	if err != nil {
		return result, err
	}

	// Read status and find out interested process
	for _, f := range procStatusFiles {
		stat, err := linuxproc.ReadProcessStat(f)
		if err != nil {
			continue
		}

		var cmd = stat.Comm[1 : len(stat.Comm)-1] // name: (cmd)
		if limit, ok := InterestedProcNames[cmd]; ok {
			result = append(result, &InterestedProc{
				StatFilePath:         f,
				Name:                 cmd,
				Pid:                  stat.Pid,
				CPULimitAsGloabl:     limit.CPULimitAsGloabl,
				CPULimitAsSingleCore: limit.CPULimitAsSingleCore,
				TotalTime:            stat.Utime + stat.Stime, // Time in user space + Time in kernal space
			})
		}
	}

	return result, nil
}

func getSystemCPUTime(stat linuxproc.CPUStat) (idleTime uint64, totalTime uint64) {
	return stat.Idle, getTotalTime(stat)
}

func getMemPercentage(memAvailable uint64, memTotal uint64) float64 {
	return 100 - (float64(100*memAvailable) / float64(memTotal))
}

func getSystemCPUPercentage(deltaSystemIdleTime uint64, deltaSystemTime uint64) float64 {
	return 100 - (float64(100*(deltaSystemIdleTime)) / float64(deltaSystemTime))
}

func getProcessCPUPercentageAsGlobal(deltaProcessCPUTime uint64, deltaSystemCPUTime uint64) float64 {
	return 100 * float64(deltaProcessCPUTime) / float64(deltaSystemCPUTime)
}

func getProcessCPUPercentageAsSingleCore(deltaProcessCPUTime uint64, deltaRealTimeInSeconds float64) float64 {
	return 100 * float64(deltaProcessCPUTime) / deltaRealTimeInSeconds / ClkTck // deltaCPUTime / ClrTck = deltaProcessCPUTime in seconds
}
