package main

import "github.com/docker/docker/client"

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
	certificate.VerifyCertificate()

	hera.Listen()
}
