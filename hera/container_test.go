package main

import (
	"testing"

	"github.com/spf13/afero"
)

var container = &Container{
	ID:       "f5020368c5bf",
	Hostname: "hostname",
	Labels:   map[string]string{},
}

func TestGetHostname(t *testing.T) {
	_, err := container.getHostname()
	if err == nil {
		t.Error("want error")
	}

	expected := "a.host.com"
	container.Labels = map[string]string{
		"hera.hostname": expected,
	}

	hostname, err := container.getHostname()
	if err != nil {
		t.Error(err)
	}

	if hostname != expected {
		t.Errorf("Unexpected hostname: %s", hostname)
	}
}

func TestGetPort(t *testing.T) {
	_, err := container.getPort()
	if err == nil {
		t.Error("want error")
	}

	expected := "8080"
	container.Labels = map[string]string{
		"hera.port": expected,
	}

	port, err := container.getPort()
	if err != nil {
		t.Error(err)
	}

	if port != expected {
		t.Errorf("Unexpected port: %s", port)
	}
}

func TestGetCertificateUsesDefaultCert(t *testing.T) {
	cert, err := container.getCertificate()
	if err != nil {
		t.Error(err)
	}

	if cert == nil {
		t.Error("Expected certificate")
	}
}

func TestGetCertificateNonexistent(t *testing.T) {
	certname := "mysite.com.pem"
	container.Labels = map[string]string{
		"hera.certificate": certname,
	}

	_, err := container.getCertificate()
	if err == nil {
		t.Error("want error")
	}
}

func TestGetCertificateExistingCustom(t *testing.T) {
	certname := "mysite.com.pem"
	cert := NewCertificate(certname)
	err := afero.WriteFile(fs, cert.fullPath(), []byte(""), 0644)
	if err != nil {
		t.Error(err)
	}

	foundcert, err := container.getCertificate()
	if err != nil {
		t.Error(err)
	}

	if foundcert.Name != certname {
		t.Errorf("Unexpected certificate name: %s", foundcert.Name)
	}
}

func TestGetCertificateMatchesOnHostname(t *testing.T) {
	certname := "mysite.com.pem"
	cert := NewCertificate(certname)
	err := afero.WriteFile(fs, cert.fullPath(), []byte(""), 0644)
	if err != nil {
		t.Error(err)
	}

	container.Labels = map[string]string{
		"hera.hostname": "mysite.com",
	}

	foundcert, err := container.getCertificate()
	if err != nil {
		t.Error(err)
	}

	if foundcert == nil {
		t.Errorf("Unexpected nil certificate")
	}
}
