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
	"https://medium.com/tailwinds-navigator/kubernetes-tip-how-does-oomkilled-work-ba71b135993b",
}

const a = "root        2178    1891  5 Nov22 ?        23:59:58 kube-apiserver --advertise-address=10.1.255.128 --allow-privileged=true --authorization-mode=Node,RBAC --client-ca-file=/etc/kubernetes/pki/ca.crt --enable-admission-plugins=NodeRestriction --enable-bootstrap-token-auth=true --encryption-provider-config=/etc/kubernetes/encryption-config.yaml --etcd-cafile=/etc/kubernetes/pki/etcd/ca.crt --etcd-certfile=/etc/kubernetes/pki/apiserver-etcd-client.crt --etcd-keyfile=/etc/kubernetes/pki/apiserver-etcd-client.key --etcd-servers=https://127.0.0.1:2379 --kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt --kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname --oidc-client-id=7589a7ea-2ee3-486e-adf6-e0036f3b8000 --oidc-groups-claim=roles --oidc-groups-prefix=oidc: --oidc-issuer-url=https://sts.windows.net/72f988bf-86f1-41af-91ab-2d7cd011db47/ --oidc-username-claim=upn --oidc-username-prefix=oidc: --proxy-client-cert-file=/etc/kubernetes/pki/front-proxy-client.crt --proxy-client-key-file=/etc/kubernetes/pki/front-proxy-client.key --requestheader-allowed-names=front-proxy-client --requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt --requestheader-extra-headers-prefix=X-Remote-Extra- --requestheader-group-headers=X-Remote-Group --requestheader-username-headers=X-Remote-User --secure-port=443 --service-account-issuer=https://kubernetes.default.svc.cluster.local --service-account-key-file=/etc/kubernetes/pki/sa.pub --service-account-signing-key-file=/etc/kubernetes/pki/sa.key --service-cluster-ip-range=10.0.0.0/16,2001:4899:6900:100::/108 --tls-cert-file=/etc/kubernetes/pki/apiserver.crt --tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256 --tls-private-key-file=/etc/kubernetes/pki/apiserver.key\n"

var configRegex = regexp.MustCompile(".*--encryption-provider-config=(.*) .*")

type KMSCacheSizeChecker struct {
	onMaster   bool
	kmsEnabled bool
	cacheSize  int
}

func (c *KMSCacheSizeChecker) Name() string {
	return "KMSCacheSize"
}

func New() *KMSCacheSizeChecker {
	return &KMSCacheSizeChecker{}
}

func (c *KMSCacheSizeChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	var results []*base.CheckResult
	match := configRegex.FindStringSubmatch(a)
	for _, m := range match {
		fmt.Printf("mathc %s\n", m)
	}
	return results, nil
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
	//todo:support other os
	if !ctx.Environment.HasFlag("linux") {
		result.Description = fmt.Sprint("Skip kms check in non-linux os")
		return result, nil
	}
	err := c.ParseCurrentLimit()
	if err != nil {
		return result, err
	}
	if !c.onMaster {
		result.Description = fmt.Sprint("Skip kms check because on non-master node ")
		return result, nil
	}
	if !c.kmsEnabled || c.cacheSize == 0 {
		result.Description = fmt.Sprint("Skip kms check because no kms cache limit found ")
		return result, nil
	}
	secretCount, err := c.getCurrentSecretsCount(ctx)
	if err != nil {
		return nil, err
	} else {
		result.Error = fmt.Sprintf("KMS cache is not enough")
		result.Description = fmt.Sprintf("Current secrets:%d, cache size:%d", secretCount, c.cacheSize)
		result.HelpLinks = helpLink
	}
	return result, nil
}

func (c *KMSCacheSizeChecker) ParseCurrentLimit() error {
	cmd := exec.Command("ps", "-ef")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return err
	} else {
		result := out.String()
		pss := strings.Split(result, "\n")
		for _, ps := range pss {
			fields := strings.Fields(ps)
			if len(fields) >= 8 && strings.ToLower(fields[7]) == "kube-apiserver" {
				c.onMaster = true
				match := configRegex.FindStringSubmatch(ps)
				if len(match) == 1 {
					configPath := match[0]
					limit, err := extractLimit(configPath)
					if err != nil {
						return err
					}
					c.kmsEnabled = true
					c.cacheSize = limit
				} else {
					fmt.Printf("Cant match")
				}
			}
		}
	}
	return nil
}

func extractLimit(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
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
