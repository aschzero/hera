package main

import (
	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

func main() {
	InitLogger()

	log.Infof("Hera v%s has started", CurrentVersion())

	client, err := NewClient()
	if err != nil {
		log.Errorf("Error connecting to the Docker daemon: %s", err)
		return
	}

	hera := &Hera{
		Client:            client,
		RegisteredTunnels: make(map[string]*Tunnel),
	}

	hera.CheckCertificates()
	hera.Revive()
	hera.Listen()
}
