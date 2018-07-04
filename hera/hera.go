package main

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
)

// Hera holds an instantiated Client and a map of registered tunnels
type Hera struct {
	Client            *Client
	RegisteredTunnels map[string]*Tunnel
}

// CheckCertificates verifies the presence of at least one cert file
func (h Hera) CheckCertificates() {
	certificateConfig := NewCertificateConfig()
	certs, err := certificateConfig.scan()
	if err != nil {
		log.Errorf("Error while checking certificates: %s", err)
		return
	}

	if len(certs) == 0 {
		log.Error(CertificateIsNeededMessage)
		return
	}

	for _, cert := range certs {
		log.Infof("Found certificate: %s", cert.Name())
	}
}

// Revive starts tunnels for containers already running
func (h Hera) Revive() {
	containers, err := h.Client.Docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Error(err)
		return
	}

	for _, c := range containers {
		container, err := NewContainer(h.Client, c.ID)
		if err != nil {
			log.Error(err)
			continue
		}

		tunnel, err := container.tryTunnel()
		if err != nil {
			continue
		}

		if err := tunnel.start(); err != nil {
			log.Errorf("Error starting tunnel: %s", err)
			continue
		}

		h.RegisterTunnel(container.ID, tunnel)
	}
}

// Listen continuously listens for container start or die events
func (h Hera) Listen() {
	log.Info("Hera is listening")

	messages, errs := h.Client.Docker.Events(context.Background(), types.EventsOptions{})

	for {
		select {
		case err := <-errs:
			if err != nil && err != io.EOF {
				log.Error(err)
			}

			os.Exit(1)

		case event := <-messages:
			if event.Status == "start" {
				h.HandleStartEvent(event)
				continue
			}

			if event.Status == "die" {
				h.HandleDieEvent(event)
				continue
			}
		}
	}
}

// HandleStartEvent tries to start a tunnel for a new container
func (h Hera) HandleStartEvent(event events.Message) {
	container, err := NewContainer(h.Client, event.ID)
	if err != nil {
		log.Error(err)
		return
	}

	tunnel, err := container.tryTunnel()
	if err != nil {
		log.Infof("Ignoring container %s: %s", container.ID, err)
		return
	}

	if err := tunnel.start(); err != nil {
		log.Errorf("Error starting tunnel: %s", err)
		return
	}

	h.RegisterTunnel(container.ID, tunnel)
}

// HandleDieEvent tries to stop a tunnel when a container is stopped
func (h Hera) HandleDieEvent(event events.Message) {
	container, err := NewContainer(h.Client, event.ID)
	if err != nil {
		log.Error(err)
		return
	}

	if tunnel, ok := h.RegisteredTunnels[container.ID]; ok {
		tunnel.stop()
	}
}

// RegisterTunnel stores a Tunnel in memory for later reference
func (h Hera) RegisterTunnel(id string, tunnel *Tunnel) {
	h.RegisteredTunnels[id] = tunnel
}
