package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

var tunnel = NewTunnel("172.21.0.3", "f56540dbf360", "host.name", "8080", NewCertificate("cert.pem"))

func TestPrepareService(t *testing.T) {
	fs = afero.NewMemMapFs()

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

	if err := tunnel.generateRunFile(); err != nil {
		t.Error(err)
	}

	_, err := fs.Stat(filepath.Join(tunnel.TunnelConfig.ServicePath, "run"))
	if os.IsNotExist(err) {
		t.Error(err)
	}
}
