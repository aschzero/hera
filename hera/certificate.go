package main

import (
	"path/filepath"
	"strings"

	tld "github.com/jpillora/go-tld"
	"github.com/spf13/afero"
)

type Certificate struct {
	Name              string
	CertificateConfig *CertificateConfig
}

type CertificateConfig struct {
	Path        string
	DefaultName string
}

func NewCertificateConfig() *CertificateConfig {
	config := &CertificateConfig{
		Path:        "/root/.cloudflared",
		DefaultName: "cert.pem",
	}

	return config
}

func NewCertificate(name string) *Certificate {
	config := NewCertificateConfig()

	if name == "" {
		name = config.DefaultName
	}

	cert := &Certificate{
		Name:              name,
		CertificateConfig: config,
	}

	return cert
}

func (c CertificateConfig) scan() ([]*Certificate, error) {
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

func (c Certificate) matchesDomain(domain string) bool {
	baseCertName := strings.Split(c.Name, ".pem")[0]

	if !strings.HasPrefix(domain, "http") {
		domain = "https://" + domain
	}

	parsed, err := tld.Parse(domain)
	if err != nil {
		return false
	}

	baseDomain := parsed.Domain + "." + parsed.TLD
	return baseDomain == baseCertName
}

const (
	CertificateIsNeededMessage = "\n Hera is unable to run without a cloudflare certificate. To fix this issue:" +
		"\n\n 1. Ensure this container has a volume mapped to `/root/.cloudflared`" +
		"\n 2. Obtain a certificate by visiting https://www.cloudflare.com/a/warp" +
		"\n 3. Rename the certificate to `cert.pem` and move it to the volume" +
		"\n\n See https://github.com/aschzero/hera#obtain-a-certificate for more info." +
		"\n\n Hera is now watching for a `cert.pem` file and will resume operation when a certificate is found.\n"
)
