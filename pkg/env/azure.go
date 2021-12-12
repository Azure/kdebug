package env

import (
	"net/http"
	"os"
	"time"
)

const (
	AzureIMDSEndpoint = "http://169.254.169.254/metadata"
)

func getAzureFlags() []string {
	// IDMS should exist on Azure VMs
	client := &http.Client{
		Timeout: time.Second,
	}
	req, _ := http.NewRequest("GET", AzureIMDSEndpoint+"/instance?api-version=2017-03-01", nil)
	req.Header.Set("Metadata", "true")
	resp, err := client.Do(req)
	if err != nil {
		// Not on Azure
		return []string{}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Not 200 status, might not be on Azure
		return []string{}
	}

	// If we are on Azure, check if it's AKS
	return append([]string{"azure"}, getAksFlags()...)
}

func getAksFlags() []string {
	// Check kubernetes directory to see if it's a AKS node
	finfo, err := os.Stat("/etc/kubernetes")
	if err != nil {
		return []string{}
	}
	if !finfo.IsDir() {
		// Not a dir
		return []string{}
	}
	return []string{"aks"}
}
