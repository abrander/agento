package ssh

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/logger"
)

type (
	Ssh struct {
		Host     string `json:"host" description:"Hostname or IP adress to connect to"`
		Port     uint16 `json:"port" description:"TCP port to connect to" default:"22"`
		Username string `json:"username" description:"Username"`
	}
)

const (
	privateKeyFilename = "id_rsa"
	publicKeyFilename  = "id_rsa.pub"
)

var (
	signer    ssh.Signer
	publicKey string
	lock      sync.Mutex
)

func PublicKey() string {
	lock.Lock()
	defer lock.Unlock()

	if publicKey != "" {
		return publicKey
	}

	pemBytes, err := ioutil.ReadFile(path.Join(configuration.StateDir, privateKeyFilename))
	if err != nil {
		pemBytes, err = GenerateKey()
		if err != nil {
			logger.Error("ssh1", err.Error())
		}
	}

	// Parse private key for ssh
	signer, err = ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		logger.Error("ssh", err.Error())
	}

	// Parse private key for generating public key
	key, err := ssh.ParseRawPrivateKey(pemBytes)
	if err != nil {
		logger.Error("ssh", err.Error())
		return ""
	}

	// Convert to RSA key
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		logger.Error("ssh", "Wrong key file format")
		return ""
	}

	// Generate public key (this is deterministic)
	rsaPubKey, err := ssh.NewPublicKey(&rsaKey.PublicKey)
	if err != nil {
		logger.Error("ssh", err.Error())
		return ""
	}

	// Generate nice string for ~/.ssh/authorized_keys
	publicKey = string(bytes.TrimSpace(ssh.MarshalAuthorizedKey(rsaPubKey))) + " https://github.com/abrander/agento\n"

	// Write file for convenience and automation
	err = ioutil.WriteFile(path.Join(configuration.StateDir, publicKeyFilename), []byte(publicKey), 0644)
	if err != nil {
		logger.Error("ssh", err.Error())
	}

	return publicKey
}

// Generate a new private RSA key for ssh
func GenerateKey() ([]byte, error) {
	logger.Green("ssh", "Generating new key")
	var pemBuffer bytes.Buffer

	// Generate RSA keypair
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Encode RSA private key to pem
	pem.Encode(&pemBuffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})

	err = ioutil.WriteFile(path.Join(configuration.StateDir, privateKeyFilename), pemBuffer.Bytes(), 0600)
	if err != nil {
		return nil, err
	}

	return pemBuffer.Bytes(), nil
}

// Connect to a remote ssh server using public key authentication
func (s *Ssh) Connect() (*ssh.Client, error) {
	dialString := fmt.Sprintf("%s:%d", s.Host, s.Port)
	logger.Yellow("ssh", "Connecting to %s as %s", dialString, s.Username)

	// We have to call PublicKey() to make sure signer is initialized
	PublicKey()

	config := &ssh.ClientConfig{
		User: s.Username,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}
	client, err := ssh.Dial("tcp", dialString, config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
