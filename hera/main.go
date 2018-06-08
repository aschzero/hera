package main

import (
	"github.com/docker/docker/client"
	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

func main() {
	InitLogger()

	log.Infof("Hera v%s has started", CurrentVersion())

	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, nil)
	if err != nil {
		log.Errorf("Error connecting to the Docker daemon: %s", err)
		return
	}

	hera := &Hera{
		Client:            cli,
		RegisteredTunnels: make(map[string]*Tunnel),
	}

	certificate := NewCertificate()
	if err := certificate.VerifyCertificate(); err != nil {
		log.Info(CertificateIsNeededMessage)
		certificate.Wait()
	}

	hera.Revive()

	hera.Listen()
}
