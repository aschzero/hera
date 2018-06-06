package main

import (
	"os"
	"testing"

	"github.com/spf13/afero"
)

var (
	ContainerHostname = "0.0.0.0"
	HeraHostname      = "host.name"
	HeraPort          = "80"
)

func TestPrepareService(t *testing.T) {
	fs = afero.NewMemMapFs()
	tunnel := NewTunnel(ContainerHostname, HeraHostname, HeraPort)

	if err := tunnel.PrepareService(); err != nil {
		t.Error(err)
	}

	_, err := fs.Stat(tunnel.TunnelConfig.ServicePath)
	if os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestGenerateConfigFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	tunnel := NewTunnel(ContainerHostname, HeraHostname, HeraPort)

	if err := tunnel.GenerateConfigFile(); err != nil {
		t.Error(err)
	}

	_, err := fs.Stat(tunnel.TunnelConfig.ConfigFilePath)
	if os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestGenerateRunFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	tunnel := NewTunnel(ContainerHostname, HeraHostname, HeraPort)

	if err := tunnel.GenerateRunFile(); err != nil {
		t.Error(err)
	}

	_, err := fs.Stat(tunnel.TunnelConfig.RunFilePath)
	if os.IsNotExist(err) {
		t.Error(err)
	}
}
