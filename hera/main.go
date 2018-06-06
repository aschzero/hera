package main

import (
	"github.com/docker/docker/client"
	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

func main() {
	InitLogger()

	log.Infof("\nHera v%s", CurrentVersion())

	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, nil)
	if err != nil {
		log.Errorf("Error when trying to connect to the Docker daemon: %s", err)
		return
	}

	hera := &Hera{
		Client:        cli,
		ActiveTunnels: make(map[string]*Tunnel),
	}

	certificate := NewCertificate()
	if err := certificate.VerifyCertificate(); err != nil {
		log.Info(CertificateIsNeededMessage)
		certificate.Wait()
	}

	hera.Listen()
}
