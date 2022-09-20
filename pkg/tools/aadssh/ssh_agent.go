package aadssh

import (
	"crypto/rsa"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// addSSHKeyToAgent adds SSH key to SSH agent
// sockPath can be a unix socket on Unix or a named pipe on Windows
func addSSHKeyToAgent(
	sockPath string,
	sshPrivKey *rsa.PrivateKey,
	sshCert *ssh.Certificate) error {

	conn, err := dialSSHAgent(sockPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	lifeTimeSecs := uint32(uint64(time.Now().Unix()) - sshCert.ValidBefore)

	client := agent.NewClient(conn)
	return client.Add(agent.AddedKey{
		Comment:      "AAD SSH Key",
		PrivateKey:   sshPrivKey,
		Certificate:  sshCert,
		LifetimeSecs: lifeTimeSecs,
	})
}
