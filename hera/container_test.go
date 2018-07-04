package main

import (
	"testing"

	"github.com/spf13/afero"
)

var container = &Container{
	ID:       "f5020368c5bf",
	Hostname: "hostname.com",
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

func TestGetCertificate(t *testing.T) {
	fs = afero.NewMemMapFs()
	cert := NewCertificate(container.Hostname)

	fs.Create(cert.fullPath())

	container.Labels = map[string]string{
		"hera.hostname": container.Hostname,
	}

	foundCert, err := container.getCertificate()
	if err != nil {
		t.Error(err)
	}

	if foundCert == nil {
		t.Error("Expected to find certificate")
	}
}

func TestGetRootHost(t *testing.T) {
	hostname := "mydomain.com"
	container.Labels = map[string]string{
		"hera.hostname": hostname,
	}

	root, err := container.getRootHost()
	if err != nil {
		t.Error(err)
	}

	if root != hostname {
		t.Errorf("Unexpected root host: %s", root)
	}

	container.Labels = map[string]string{
		"hera.hostname": "sub.mysite.co.uk",
	}

	root, err = container.getRootHost()
	if err != nil {
		t.Error(err)
	}

	if root != "mysite.co.uk" {
		t.Errorf("Unexpected root host: %s", root)
	}
}
