package main

import (
	"testing"

	"github.com/spf13/afero"
)

func newTunnel() *Tunnel {
	config := &TunnelConfig{
		IP:             "172.23.0.4",
		Hostname:       "f56540dbf360",
		TunnelHostname: "site.tld",
		TunnelPort:     "80",
	}
	cert := NewDefaultCertificate()

	return NewTunnel(config, cert)
}

func TestWriteConfigFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	tunnel := newTunnel()

	err := tunnel.writeConfigFile()
	if err != nil {
		t.Error(err)
	}

	exists, err := afero.Exists(fs, tunnel.Service.configFilePath())
	if err != nil {
		t.Error(err)
	}

	if !exists {
		t.Error("Expected config to exist")
	}
}

func TestWriteRunFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	tunnel := newTunnel()

	err := tunnel.writeRunFile()
	if err != nil {
		t.Error(err)
	}

	exists, err := afero.Exists(fs, tunnel.Service.runFilePath())
	if err != nil {
		t.Error(err)
	}

	if !exists {
		t.Error("Expected run to exist")
	}
}
