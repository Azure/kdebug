package kmscachesize

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Azure/kdebug/pkg/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var helpLink = []string{
	"https://kubernetes.io/docs/tasks/administer-cluster/kms-provider/#configuring-the-kms-provider-kms-v2",
}

const cacheSizeAlertThreshold = 0.8

var configRegex = regexp.MustCompile(".*--encryption-provider-config=(\\S*) .*")

type KMSCacheSizeChecker struct {
}

func (c *KMSCacheSizeChecker) Name() string {
	return "KMSCacheSize"
}

func New() *KMSCacheSizeChecker {
	return &KMSCacheSizeChecker{}
}

func (c *KMSCacheSizeChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	var results []*base.CheckResult
	cacheSizeResult, err := c.checkCacheSizeState(ctx)
	if err != nil {
		return nil, err
	}
	results = append(results, cacheSizeResult)
	return results, nil
}

func (c *KMSCacheSizeChecker) checkCacheSizeState(ctx *base.CheckContext) (*base.CheckResult, error) {
	result := &base.CheckResult{
		Checker: c.Name(),
	}
	if !ctx.Environment.HasFlag("linux") {
		result.Description = fmt.Sprint("Skip kms check in non-linux os")
		return result, nil
	}

	cacheSize, err := parseCurrentLimit()
	if err != nil {
		return result, err
	}
	if cacheSize == 0 {
		result.Description = fmt.Sprint("Skip kms check because no kms cache limit found ")
		return result, nil
	}

	secretCount, err := c.getCurrentSecretsCount(ctx)
	if err != nil {
		return nil, err
	}
	result.Description = fmt.Sprintf("Current secrets:%d, cache size:%d", secretCount, cacheSize)

	if float32(secretCount) > (float32(cacheSize) * cacheSizeAlertThreshold) {
		result.Error = fmt.Sprintf("KMS cache is not enough")
		result.HelpLinks = helpLink
	}

	return result, nil
}

func parseCurrentLimit() (int, error) {
	if configFilePath, err := getKmsConfigFile(); err != nil || configFilePath == "" {
		return 0, err
	} else {
		return extractLimit(configFilePath)
	}
}

func getKmsConfigFile() (string, error) {
	cmd := exec.Command("ps", "-ef")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", err
	} else {
		result := out.String()
		pss := strings.Split(result, "\n")
		for _, ps := range pss {
			fields := strings.Fields(ps)
			if len(fields) >= 8 && strings.ToLower(fields[7]) == "kube-apiserver" {
				match := configRegex.FindStringSubmatch(ps)
				if len(match) == 2 {
					configPath := match[1]
					return configPath, nil
				}
			}
		}
	}
	return "", nil
}

func extractLimit(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "cachesize") {
			fileds := strings.Fields(line)
			if len(fileds) == 2 {
				limit, err := strconv.Atoi(fileds[1])
				return limit, err
			}
		}
	}
	return 0, nil
}

func (c *KMSCacheSizeChecker) getCurrentSecretsCount(ctx *base.CheckContext) (int, error) {
	if ctx.KubeClient == nil {
		return 0, errors.New("no available kube client")
	}
	client := ctx.KubeClient
	secrets, err := client.CoreV1().Secrets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return 0, err
	}
	return len(secrets.Items), nil
}
