package diskusage

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"os/exec"

	"github.com/Azure/kdebug/pkg/base"
)

const (
	NoHighDiskUsageResult        = "Disk usage is in normal size. No additional action required."
	HighUsageRecommandation      = "Run bash command: truncate -s 0 /path/to/file to reduce disk usage."
	FailedToCheckDiskUsageWithDf = "Failed to check disk usage with 'df -h'"
)

var (
	DfHeaders = []string{
		"Filesystem",
		"Size",
		"Used",
		"Avail",
		"Use%",
		"Mounted",
		"on",
	}

	DiskUsageRateThreshold = 70
	InterestedBigFilePath  = []string{
		"/var/log",
	}
	InterestedBigFileNum = 10

	HighdfRecommandations = []string{HighUsageRecommandation}
)

type DfRow struct {
	Filesystem string
	Size       string
	Used       string
	Avail      string
	Use        int
	MountedOn  string
}

type DiskUsageChecker struct {
}

func New() *DiskUsageChecker {
	return &DiskUsageChecker{}
}

func (c *DiskUsageChecker) Name() string {
	return "DiskUsage"
}

func (c *DiskUsageChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	result := []*base.CheckResult{}

	rst, err := c.getDiskUsage()
	if err != nil {
		return result, err
	}
	result = append(result, rst)

	return result, nil
}

func (c *DiskUsageChecker) getDiskUsage() (*base.CheckResult, error) {
	out, err := exec.Command("df", "-h").Output()
	if err != nil {
		return &base.CheckResult{
			Checker:     c.Name(),
			Error:       FailedToCheckDiskUsageWithDf,
			Description: err.Error(),
		}, nil
	}

	rows, err := parseDfResult(string(out))
	if err != nil {
		return &base.CheckResult{
			Checker:     c.Name(),
			Error:       FailedToCheckDiskUsageWithDf,
			Description: err.Error(),
		}, nil
	}

	found, row := getUsageAt("/", rows)
	if found && row.Use > DiskUsageRateThreshold {
		bigFiles := []string{}

		for _, path := range InterestedBigFilePath {
			output, err := FindTopSizeFiles(path, InterestedBigFileNum)
			if err != nil {
				return &base.CheckResult{
					Checker:         c.Name(),
					Description:     FormatHighDfDescription(row),
					Error:           err.Error(),
					Recommendations: HighdfRecommandations,
				}, nil
			}

			bigFiles = append(bigFiles, output)
		}

		return &base.CheckResult{
			Checker:         c.Name(),
			Error:           "Disk is reaching high usage. Details: " + FormatHighDfDescription(row),
			Description:     "\n" + strings.Join(bigFiles, "\n"),
			Recommendations: HighdfRecommandations,
		}, nil
	}

	return &base.CheckResult{
		Checker:     c.Name(),
		Description: NoHighDiskUsageResult,
	}, nil
}

func getUsageAt(path string, rows []DfRow) (bool, DfRow) {
	for _, row := range rows {
		if row.MountedOn == path {
			return true, row
		}
	}

	return false, DfRow{}
}

func parseDfResult(output string) ([]DfRow, error) {
	lines := strings.Split(output, "\n")
	result := make([]DfRow, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		ds := strings.Fields(strings.TrimSpace(line))
		if ds[0] == DfHeaders[0] {
			// header
			if !reflect.DeepEqual(ds, DfHeaders) {
				return result, errors.New(fmt.Sprintf("Result in df has wrong header format. Expected %v, Actually %v", DfHeaders, ds))
			}
			continue
		}

		row, err := parseDfRow(ds)
		if err != nil {
			return nil, err
		}

		result = append(result, row)
	}

	return result, nil
}

func parseDfRow(row []string) (DfRow, error) {
	if len(row) != len(DfHeaders)-1 {
		return DfRow{}, fmt.Errorf(`unexpected row column number %v (expected %v)`, row, DfHeaders)
	}

	return DfRow{
		Filesystem: strings.TrimSpace(row[0]),
		Size:       strings.TrimSpace(row[1]),
		Used:       strings.TrimSpace(row[2]),
		Avail:      strings.TrimSpace(row[3]),
		Use:        AtoiHepler(strings.TrimSpace(strings.Replace(row[4], "%", "", -1))),
		MountedOn:  strings.TrimSpace(row[5]),
	}, nil
}

func AtoiHepler(s string) int {
	rst, _ := strconv.Atoi(s)
	return rst
}

func FormatHighDfDescription(row DfRow) string {
	return fmt.Sprintf("[Used %d%%] Filesystem: %s, UsedSize: %s, AvailableSize: %s, MountedOn %s", row.Use, row.Filesystem, row.Used, row.Avail, row.MountedOn)
}

func FindTopSizeFiles(path string, topCount int) (string, error) {
	commandline := fmt.Sprintf("du -ah %s | sort -rh | head -n %d", path, topCount)
	out, err := exec.Command("bash", "-c", commandline).Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}
