package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
)

const KubernetesServiceHost = "KUBERNETES_SERVICE_HOST"

func getK8sFlags() []string {
	var flags []string
	if inK8s() {
		flags = append(flags, "k8s")
	}
	return flags
}

func inK8s() bool {
	//check if in a pod
	for _, e := range os.Environ() {
		if strings.Contains(e, KubernetesServiceHost) {
			return true
		}
	}
	// check in a host vm
	processes, err := process.Processes()
	if err != nil {
		log.Warn(fmt.Sprintf("List process error %v\n", err))
		return false
	} else {
		for _, proc := range processes {
			name, err := proc.Name()
			if err != nil {
				log.Warn(fmt.Sprintf("List process error %v. Skip in-cluster tcp checking\n", err))
				return false
			}
			if name == "kubelet" {
				return true
			}
		}
	}
	return false
}
