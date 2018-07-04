package main

import (
	"os"
	"testing"

	"github.com/spf13/afero"
)

var (
	containerHostname = "0.0.0.0"
	heraHostname      = "host.name"
	heraPort          = "80"
	certificate       = NewCertificate("cert.pem")
)

func TestPrepareService(t *testing.T) {
	fs = afero.NewMemMapFs()
	tunnel := NewTunnel(containerHostname, heraHostname, heraPort, certificate)

	if err := tunnel.prepareService(); err != nil {
		t.Error(err)
	}

	_, err := fs.Stat(tunnel.TunnelConfig.ServicePath)
	if os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestGenerateConfigFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	tunnel := NewTunnel(containerHostname, heraHostname, heraPort, certificate)

	if err := tunnel.generateConfigFile(); err != nil {
		t.Error(err)
	}

	_, err := fs.Stat(tunnel.TunnelConfig.ConfigFilePath)
	if os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestGenerateRunFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	tunnel := NewTunnel(containerHostname, heraHostname, heraPort, certificate)

	if err := tunnel.generateRunFile(); err != nil {
		t.Error(err)
	}

	_, err := fs.Stat(tunnel.TunnelConfig.RunFilePath)
	if os.IsNotExist(err) {
		t.Error(err)
	}
}
