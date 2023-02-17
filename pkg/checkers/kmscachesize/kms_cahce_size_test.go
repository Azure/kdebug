package kmscachesize

import (
	"fmt"
	"testing"
)

func TestKmsConfigPath(t *testing.T) {
	const ps = "root        2178    1891  5 Nov22 ?        23:59:58 kube-apiserver --advertise-address=10.1.255.128 --allow-privileged=true --authorization-mode=Node,RBAC --client-ca-file=/etc/kubernetes/pki/ca.crt --enable-admission-plugins=NodeRestriction --enable-bootstrap-token-auth=true --encryption-provider-config=/etc/kubernetes/encryption-config.yaml --etcd-cafile=/etc/kubernetes/pki/etcd/ca.crt --etcd-certfile=/etc/kubernetes/pki/apiserver-etcd-client.crt --etcd-keyfile=/etc/kubernetes/pki/apiserver-etcd-client.key --etcd-servers=https://127.0.0.1:2379 --kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt --kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname --oidc-client-id=7589a7ea-2ee3-486e-adf6-e0036f3b8000 --oidc-groups-claim=roles --oidc-groups-prefix=oidc: --oidc-issuer-url=https://sts.windows.net/72f988bf-86f1-41af-91ab-2d7cd011db47/ --oidc-username-claim=upn --oidc-username-prefix=oidc: --proxy-client-cert-file=/etc/kubernetes/pki/front-proxy-client.crt --proxy-client-key-file=/etc/kubernetes/pki/front-proxy-client.key --requestheader-allowed-names=front-proxy-client --requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt --requestheader-extra-headers-prefix=X-Remote-Extra- --requestheader-group-headers=X-Remote-Group --requestheader-username-headers=X-Remote-User --secure-port=443 --service-account-issuer=https://kubernetes.default.svc.cluster.local --service-account-key-file=/etc/kubernetes/pki/sa.pub --service-account-signing-key-file=/etc/kubernetes/pki/sa.key --service-cluster-ip-range=10.0.0.0/16,2001:4899:6900:100::/108 --tls-cert-file=/etc/kubernetes/pki/apiserver.crt --tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256 --tls-private-key-file=/etc/kubernetes/pki/apiserver.key\n"
	match := configRegex.FindStringSubmatch(ps)
	if len(match) != 2 {
		t.Error("Can't parse Kms config")
	}
	if match[1] != "/etc/kubernetes/encryption-config.yaml" {
		t.Error(fmt.Sprintf("Parsed wrong config file path:%s", match[1]))
	}
}
