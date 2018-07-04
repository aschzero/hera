package main

import (
	"testing"

	"github.com/spf13/afero"
)

var (
	certPath = "/root/.cloudflared"
	certName = "cert.pem"
)

func TestNewCertificateContainsDefaultName(t *testing.T) {
	cert := NewCertificate("")

	if cert.Name != cert.CertificateConfig.DefaultName {
		t.Errorf("Got unexpected certificate name: %s", cert.Name)
	}
}

func TestCertificateContainsCustomName(t *testing.T) {
	name := "mysite.com.pem"
	cert := NewCertificate(name)

	if cert.Name != name {
		t.Errorf("Got unexpected certificate name: %s", cert.Name)
	}
}

func TestFullPathContainsPath(t *testing.T) {
	cert := NewCertificate("cert.pem")
	expected := "/root/.cloudflared/cert.pem"

	if cert.fullPath() != expected {
		t.Errorf("Got unexpected certificate path: %s", cert.Name)
	}
}

func TestIsExist(t *testing.T) {
	fs = afero.NewMemMapFs()
	cert := NewCertificate("cert.pem")

	err := afero.WriteFile(fs, cert.fullPath(), []byte(""), 0644)
	if err != nil {
		t.Error(err)
	}

	exists := cert.isExist()
	if !exists {
		t.Error("Certificate file does not exist")
	}
}

func TestMatchesDomain(t *testing.T) {
	cert := NewCertificate("hostname.com.pem")

	matches := cert.matchesDomain("hostname.com")
	if !matches {
		t.Errorf("Expected cert and domain names to match")
	}

	matches = cert.matchesDomain("sub.hostname.com")
	if !matches {
		t.Errorf("Expected cert and domain names to match")
	}

	matches = cert.matchesDomain("horsename.com")
	if matches {
		t.Errorf("Expected cert and domain names to not match")
	}
}
