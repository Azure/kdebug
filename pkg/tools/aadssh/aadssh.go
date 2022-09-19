package aadssh

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/Azure/kdebug/pkg/base"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	// Extracted from Azure CLI code
	AzureCLIClientId           = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"
	AzureCLIDirName            = ".azure"
	AzureCLITokenCacheFileName = "msal_token_cache.json"
	SSHDirName                 = ".aad-ssh"
	SSHPrivateKeyName          = "id_rsa"
	SSHPublicKeyName           = "id_rsa.pub"
	SSHCertificateName         = "id_rsa-cert.pub"
)

type AadSsh struct {
}

func New() *AadSsh {
	return &AadSsh{}
}

func (c *AadSsh) Name() string {
	return "AAD SSH"
}

func (c *AadSsh) Run(ctx *base.ToolContext) error {
	if ctx.AadSsh.Cloud == "" {
		// Default to public cloud
		ctx.AadSsh.Cloud = "azurecloud"
	}

	// Ensure key dir
	sshDir, err := ensureSSHKeyDir(SSHDirName)
	if err != nil {
		return fmt.Errorf("Fail to ensure SSH directory: %+v", err)
	}

	// Load SSH private key
	sshPrivKeyPath := path.Join(sshDir, SSHPrivateKeyName)
	sshPrivKey, err := createOrLoadSSHPrivateKey(sshPrivKeyPath)
	if err != nil {
		return fmt.Errorf("Fail to create or load SSH private key: %+v", err)
	}
	log.WithFields(log.Fields{"path": sshPrivKeyPath}).Info("Loaded SSH private key")

	// Save SSH public key
	sshPubKey, err := ssh.NewPublicKey(&sshPrivKey.PublicKey)
	if err != nil {
		return fmt.Errorf("Fail to create SSH public key: %+v", err)
	}
	sshPubKeyPath := path.Join(sshDir, SSHPublicKeyName)
	if err = saveSSHPublicKey(sshPubKey, sshPubKeyPath); err != nil {
		return fmt.Errorf("Fail to save SSH public key: %+v", err)
	}
	log.WithFields(log.Fields{"path": sshPubKeyPath}).Info("Saved SSH public key")

	// Try existing certificate
	sshCertPath := path.Join(sshDir, SSHCertificateName)
	sshCert, err := loadSSHCertificate(sshCertPath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("Fail to load existing SSH certificate")
		log.Info("Acquire a new SSH certificate from AAD")

		// Acquire a certificate from AAD
		sshCert, err = acquireCertificate(ctx.AadSsh.Cloud, ctx.AadSsh.UseAzureCLI, sshPubKey)
		if err != nil {
			return fmt.Errorf("Fail to acquire SSH certificate from AAD: %+v", err)
		}

		// Save SSH certificate to file
		sshCertContent := ssh.CertAlgoRSAv01 + " " + base64.StdEncoding.EncodeToString(sshCert.Marshal())
		if err = saveSSHCertificate(sshCertContent, sshCertPath); err != nil {
			return fmt.Errorf("Fail to save SSH certificate: %+v", err)
		}
		log.WithFields(log.Fields{"path": sshCertPath}).Info("Saved SSH certificate")
	} else {
		log.WithFields(log.Fields{"path": sshCertPath}).Info("Loaded valid SSH certificate")
	}

	// Add SSH key to SSH agent
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock != "" {
		if err = addSSHKeyToAgent(sshAuthSock, sshPrivKey, sshCert); err != nil {
			return fmt.Errorf("Fail to add SSH key to agent: %+v", err)
		}
		log.WithFields(log.Fields{"path": sshPrivKeyPath}).Info("Added SSH key to agent")
	}

	// Call SSH there are remaining args
	if len(ctx.Args) > 0 {
		args := getSSHArgs(ctx.Args, sshPrivKeyPath, sshAuthSock != "")
		log.WithFields(log.Fields{"args": args}).Info("Starting SSH")
		cmd := exec.Command("ssh", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Start(); err != nil {
			return fmt.Errorf("Fail to start SSH: %+v", err)
		}
		cmd.Wait()
	}

	return nil
}
