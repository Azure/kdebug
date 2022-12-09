package kmscachesize

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/kdebug/pkg/base"
)

var helpLink = []string{
	"https://kubernetes.io/docs/tasks/administer-cluster/kms-provider/#configuring-the-kms-provider-kms-v2",
}

const cacheSizeAlertThreshold = 0.8
const kmsConfigCmd = "--encryption-provider-config="

type encConfig struct {
	Resources []encResource `yaml:"resources"`
}

type encResource struct {
	Providers []encProvider `yaml:"providers"`
}

type encProvider struct {
	Kms encKms `yaml:"kms"`
}

type encKms struct {
	CacheSize int `yaml:"cachesize"`
}

type KMSCacheSizeChecker struct {
}

func (c *KMSCacheSizeChecker) Name() string {
	return "KMSCacheSize"
}

func New() *KMSCacheSizeChecker {
	return &KMSCacheSizeChecker{}
}

func (c *KMSCacheSizeChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	if !ctx.Environment.HasFlag("linux") {
		log.Debugf("Skip %s checker in non-linux os", c.Name())
		return []*base.CheckResult{}, nil
	}

	if ctx.KubeClient == nil {
		log.Debugf("Skip %s checker due to no kube config provided", c.Name())
		return []*base.CheckResult{}, nil
	}

	kmsConfigPath, err := getKmsConfigPath()
	if err != nil {
		log.Debugf("Cannot find KMS config file: %s", err)
		return []*base.CheckResult{}, nil
	}

	cacheSize, err := getKmsCacheSize(kmsConfigPath)
	if err != nil {
		return nil, err
	}

	log.Debugf("KMS cache size: %d", cacheSize)

	if cacheSize == 0 {
		log.Debugf("There's no limit for KMS cache size")
		return []*base.CheckResult{}, nil
	}

	secretsCount, err := c.getCurrentSecretsCount(ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf("Secrets count: %d", secretsCount)

	result := &base.CheckResult{
		Checker:     c.Name(),
		Description: fmt.Sprintf("Current secrets:%d, cache size:%d.", secretsCount, cacheSize),
	}

	if float32(secretsCount) > (float32(cacheSize) * cacheSizeAlertThreshold) {
		result.Error = fmt.Sprintf("KMS cache size is insufficient.")
		result.Description += fmt.Sprintf(" When number of secrets exceeds KMS cache size, Kubernetes may suffer frmo significant performance issue.")
		result.HelpLinks = helpLink
	}

	return []*base.CheckResult{result}, nil
}

func getKmsConfigPath() (string, error) {
	procs, err := process.Processes()
	if err != nil {
		return "", err
	}

	for _, proc := range procs {
		procName, err := proc.Name()
		if err != nil {
			log.Errorf("Fail get proc name for pid: %d", proc.Pid)
			continue
		}

		if strings.ToLower(procName) == "kube-apiserver" {
			cmds, err := proc.CmdlineSlice()
			if err != nil {
				log.Errorf("Fail get proc cmdline for: %s", procName)
				continue
			}

			for _, cmd := range cmds {
				if strings.HasPrefix(cmd, kmsConfigCmd) {
					return cmd[len(kmsConfigCmd):], nil
				}
			}

			return "", errors.New("API server doesn't have KMS configured")
		}
	}

	return "", errors.New("Fail to find api server process")
}

func getKmsCacheSize(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	var config encConfig
	err = decoder.Decode(&config)
	if err != nil {
		return 0, err
	}

	if len(config.Resources) > 0 && len(config.Resources[0].Providers) > 0 {
		return config.Resources[0].Providers[0].Kms.CacheSize, nil
	} else {
		return 0, fmt.Errorf("Fail to parse cache size from kms config: %s", path)
	}
}

func (c *KMSCacheSizeChecker) getCurrentSecretsCount(ctx *base.CheckContext) (int, error) {
	client := ctx.KubeClient
	secrets, err := client.CoreV1().Secrets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("Fail to list secrets from Kubernetes: %s", err)
	}
	return len(secrets.Items), nil
}
