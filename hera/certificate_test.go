package main

import (
	"os"
	"testing"

	"github.com/spf13/afero"
)

var config = NewCertificateConfig()
var defaultCert = NewDefaultCertificate()

func TestNewCertificateConfig(t *testing.T) {
	if config.Path == "" {
		t.Error("Expected path")
	}
}

func TestNewCertificate(t *testing.T) {
	name := "mysite.com.pem"
	cert := NewCertificate(name)

	if cert.Name != name {
		t.Errorf("Unexpected name: %s", cert.Name)
	}
}

func TestNewDefaultCertificate(t *testing.T) {
	if defaultCert.Name != "cert.pem" {
		t.Errorf("Unexpected name: %s", defaultCert.Name)
	}
}

func TestScanAllEmpty(t *testing.T) {
	fs = afero.NewMemMapFs()
	fs.Mkdir(config.Path, os.ModeDir)

	certs, err := config.scanAll()
	if err != nil {
		t.Error(err)
	}

	if len(certs) > 0 {
		t.Error("Expected no certificates yet")
	}
}

func TestScanAllExisting(t *testing.T) {
	fs = afero.NewMemMapFs()
	fs.Mkdir(config.Path, os.ModeDir)

	certs := []*Certificate{
		NewCertificate("a.pem"),
		NewCertificate("b.pem"),
		NewCertificate("c.pem"),
	}

	for _, newCert := range certs {
		fs.Create(newCert.fullPath())
	}

	certs, err := config.scanAll()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 3 {
		t.Errorf("Unexpected scan results, got %d", len(certs))
	}
}

func TestBelongsToHost(t *testing.T) {
	cert := NewCertificate("hostname.com.pem")

	belongs := cert.belongsToHost("hostname.com")
	if !belongs {
		t.Errorf("Expected cert and host to belong")
	}

	belongs = cert.belongsToHost("horsename.com")
	if belongs {
		t.Errorf("Expected cert to not belong")
	}
}

func TestFullPath(t *testing.T) {
	cert := NewCertificate("mysite.pem")
	expected := "/root/.cloudflared/mysite.pem"

	if cert.fullPath() != expected {
		t.Errorf("Unexpected certificate path: %s", cert.Name)
	}
}

func TestIsExist(t *testing.T) {
	fs = afero.NewMemMapFs()
	exists := defaultCert.isExist()
	if exists {
		t.Error("Unexpected existing certificate")
	}

	fs.Create(defaultCert.fullPath())

	exists = defaultCert.isExist()
	if !exists {
		t.Error("Certificate file does not exist")
	}
}
