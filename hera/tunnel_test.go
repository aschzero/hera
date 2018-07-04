package main

import (
	"os"
	"testing"

	"github.com/spf13/afero"
)

var tunnel = &Tunnel{
	ContainerHostname: "0.0.0.0",
	HeraHostname:      "host.name",
	HeraPort:          "8080",
	Certificate:       NewCertificate("cert.pem"),
}

func TestPrepareService(t *testing.T) {
	fs = afero.NewMemMapFs()
	tunnel := NewTunnel(tunnel.ContainerHostname, tunnel.HeraHostname, tunnel.HeraPort, tunnel.Certificate)

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
	tunnel := NewTunnel(tunnel.ContainerHostname, tunnel.HeraHostname, tunnel.HeraPort, tunnel.Certificate)

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
	tunnel := NewTunnel(tunnel.ContainerHostname, tunnel.HeraHostname, tunnel.HeraPort, tunnel.Certificate)

	if err := tunnel.generateRunFile(); err != nil {
		t.Error(err)
	}

	_, err := fs.Stat(tunnel.TunnelConfig.RunFilePath)
	if os.IsNotExist(err) {
		t.Error(err)
	}
}
