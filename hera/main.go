package main

import (
	"flag"
	"fmt"

	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

func main() {
	version := flag.Bool("version", false, "Print current version")
	flag.Parse()

	if *version {
		fmt.Println(Version)
		return
	}

	InitLogger()
	log.Infof("Hera v%s has started", Version)

	client, err := NewClient()
	if err != nil {
		log.Errorf("Error connecting to the Docker daemon: %s", err)
		return
	}

	hera := &Hera{
		Client:            client,
		RegisteredTunnels: make(map[string]*Tunnel),
	}

	hera.checkCertificates()
	hera.revive()
	hera.listen()
}
