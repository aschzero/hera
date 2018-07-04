package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type Certificate struct {
	Name              string
	CertificateConfig *CertificateConfig
}

type CertificateConfig struct {
	Path string
}

func NewCertificateConfig() *CertificateConfig {
	config := &CertificateConfig{
		Path: "/root/.cloudflared",
	}

	return config
}

func NewCertificate(name string) *Certificate {
	config := NewCertificateConfig()

	cert := &Certificate{
		Name:              name,
		CertificateConfig: config,
	}

	return cert
}

func NewDefaultCertificate() *Certificate {
	config := NewCertificateConfig()

	cert := &Certificate{
		Name:              "cert.pem",
		CertificateConfig: config,
	}

	return cert
}

func (c CertificateConfig) scanAll() ([]*Certificate, error) {
	var certs []*Certificate

	files, err := afero.ReadDir(fs, c.Path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		cert := NewCertificate(file.Name())
		certs = append(certs, cert)
	}

	return certs, nil
}

func (c CertificateConfig) findMatchingCertificate(hostname string) (*Certificate, error) {
	certs, err := c.scanAll()
	if err != nil {
		return nil, fmt.Errorf("Unable to scan for available certificates: %s", err)
	}

	for _, cert := range certs {
		if cert.belongsToHost(hostname) {
			return cert, nil
		}
	}

	defaultCert := NewDefaultCertificate()
	log.Infof("Trying `%s` as a fallback", defaultCert.fullPath())

	if !defaultCert.isExist() {
		return nil, fmt.Errorf("Couldn't find certificate. Tried searching for both `%s` and `%s`", hostname, defaultCert.Name)
	}

	return defaultCert, nil
}

func (c Certificate) belongsToHost(host string) bool {
	baseCertName := strings.Split(c.Name, ".pem")[0]

	return host == baseCertName
}

func (c Certificate) fullPath() string {
	return filepath.Join(c.CertificateConfig.Path, c.Name)
}

func (c Certificate) isExist() bool {
	exists, err := afero.Exists(fs, c.fullPath())
	if err != nil {
		log.Error(err)
	}

	return exists
}

const (
	CertificateIsNeededMessage = "\n Hera is unable to run without a cloudflare certificate. To fix this issue:" +
		"\n\n 1. Ensure this container has a volume mapped to `/root/.cloudflared`" +
		"\n 2. Obtain a certificate by visiting https://www.cloudflare.com/a/warp" +
		"\n 3. Rename the certificate to `cert.pem` and move it to the volume" +
		"\n\n See https://github.com/aschzero/hera#obtain-a-certificate for more info." +
		"\n\n Hera is now watching for a `cert.pem` file and will resume operation when a certificate is found.\n"
)
