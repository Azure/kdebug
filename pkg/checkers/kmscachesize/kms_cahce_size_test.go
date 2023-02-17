package kmscachesize

import (
	"fmt"
	"testing"
)

func TestKmsConfigPath(t *testing.T) {
	const ps = "root        2178    1891  5 Nov22 ?        23:59:58 kube-apiserver --encryption-provider-config=/etc/kubernetes/encryption-config.yaml \n"
	match := configRegex.FindStringSubmatch(ps)
	if len(match) != 2 {
		t.Error("Can't parse Kms config")
	}
	if match[1] != "/etc/kubernetes/encryption-config.yaml" {
		t.Error(fmt.Sprintf("Parsed wrong config file path:%s", match[1]))
	}
}
