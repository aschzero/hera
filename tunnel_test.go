package main

import (
	"testing"

    "github.com/spf13/afero"
)

func newTunnel() *Tunnel {
	config := &TunnelConfig{
		IP:       "172.23.0.4",
		Hostname: "site.tld",
		Port:     "80",
	}
	cert := NewCertificate("site.tld.pem", afero.NewMemMapFs())

	return NewTunnel(config, cert)
}

func TestWriteConfigFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	tunnel := newTunnel()

	err := tunnel.writeConfigFile()
	if err != nil {
		t.Error(err)
	}

	exists, err := afero.Exists(fs, tunnel.Service.ConfigFilePath())
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

	exists, err := afero.Exists(fs, tunnel.Service.RunFilePath())
	if err != nil {
		t.Error(err)
	}

	if !exists {
		t.Error("Expected run to exist")
	}
}
