package aadssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// ensureSSHKeyDir creates a directory under user home for storing SSH keys
// returns directory path
func ensureSSHKeyDir(dirName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sshDir := path.Join(homeDir, dirName)
	if _, err = os.Stat(sshDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(sshDir, 0700); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return sshDir, nil
}

// createOrLoadSSHPrivateKey creates or loads a SSH private key from file
// returns RSA private key
func createOrLoadSSHPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	if _, err := os.Stat(keyPath); err == nil {
		f, err := os.Open(keyPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		content, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(content)
		if block == nil {
			return nil, fmt.Errorf("Empty PEM block")
		}

		return x509.ParsePKCS1PrivateKey(block.Bytes)
	} else {
		if os.IsNotExist(err) {
			key, err := rsa.GenerateKey(rand.Reader, 4096)
			if err != nil {
				return nil, err
			}

			der := x509.MarshalPKCS1PrivateKey(key)
			content := pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: der,
			})
			err = os.WriteFile(keyPath, content, 0600)
			if err != nil {
				return nil, err
			}

			return key, nil
		} else {
			return nil, err
		}
	}
}

// parseSSHPublicKey parses exponent and modulus part from SSH public key
// returns base64 encoded exponent and modulus
func parseSSHPublicKey(pubKey ssh.PublicKey) (e string, n string, err error) {
	keyBytes := pubKey.Marshal()
	// <algorithm>,<exponent>,<modulus>
	fields := [][]byte{}

	read := 0
	for read < len(keyBytes) {
		length := int(binary.BigEndian.Uint32(keyBytes[read : read+4]))
		read += 4
		fields = append(fields, keyBytes[read:read+length])
		read += length
	}

	return base64.RawURLEncoding.EncodeToString(fields[1]),
		base64.RawURLEncoding.EncodeToString(fields[2]),
		nil
}

// saveSSHPublicKey saves SSH public key to file
func saveSSHPublicKey(key ssh.PublicKey, path string) error {
	content := ssh.MarshalAuthorizedKey(key)
	return os.WriteFile(path, content, 0600)
}

// loadSSHCertificate loads SSH certificate from file
func loadSSHCertificate(path string) (*ssh.Certificate, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Fail to open SSH certificate file: %+v", err)
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Fail to read SSH certificate file: %+v", err)
	}

	parts := strings.Split(string(content), " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("SSH certificate file is in bad format")
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("Fail to decode SSH certificate: %+v", err)
	}

	pubKey, err := ssh.ParsePublicKey(data)
	if err != nil {
		return nil, fmt.Errorf("Fail to parse SSH certificate: %+v", err)
	}

	sshCert, ok := pubKey.(*ssh.Certificate)
	if !ok {
		return nil, fmt.Errorf("Not a SSH certificate")
	}

	validBefore := time.Unix(int64(sshCert.ValidBefore), 0)
	validAfter := time.Unix(int64(sshCert.ValidAfter), 0)
	valid := time.Now().Before(validBefore) && time.Now().After(validAfter)
	if !valid {
		return nil, fmt.Errorf("SSH certificate has expired. Valid before: %s. Valid after: %s",
			validBefore, validAfter)
	}

	return sshCert, nil
}

// saveSSHCertificate saves SSH certificate to file
func saveSSHCertificate(content, path string) error {
	return os.WriteFile(path, []byte(content), 0600)
}

// getSSHArgs returns command line arguments when calling SSH command
func getSSHArgs(inputArgs []string, sshPrivKeyPath string, useSSHAgent bool) []string {
	args := inputArgs
	argsMap := make(map[string]bool)
	for _, arg := range inputArgs {
		argsMap[arg] = true
	}

	if useSSHAgent && !argsMap["-A"] {
		args = append(args, "-A")
	}

	if !useSSHAgent && !argsMap["-i"] {
		args = append(args, "-i", sshPrivKeyPath)
	}

	return args
}
