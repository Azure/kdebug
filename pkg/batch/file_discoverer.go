package batch

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type FileBatchDiscoverer struct {
	Path string
}

func (d *FileBatchDiscoverer) Discover() ([]string, error) {
	var file *os.File
	if d.Path == "-" {
		file = os.Stdin
	} else {
		file, err := os.Open(d.Path)
		if err != nil {
			return nil, fmt.Errorf("Fail to open machines list file %s due to: %+v",
				d.Path, err)
		}
		defer file.Close()
	}

	var machines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := string(scanner.Text())
		line = strings.TrimSpace(line)
		machines = append(machines, line)
	}

	return machines, nil
}
