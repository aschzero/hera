package main

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
)

type Hera struct {
	Client            *Client
	RegisteredTunnels map[string]*Tunnel
}

func (h Hera) checkCertificates() {
	certificateConfig := NewCertificateConfig()
	certs, err := certificateConfig.scanAll()
	if err != nil || len(certs) == 0 {
		log.Error(CertificateIsNeededMessage)
		return
	}

	for _, cert := range certs {
		log.Infof("Found certificate: %s", cert.Name)
	}
}

func (h Hera) revive() {
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

		h.registerTunnel(container.ID, tunnel)
	}
}

func (h Hera) listen() {
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
				h.handleStartEvent(event)

				continue
			}

			if event.Status == "die" {
				h.handleDieEvent(event)
				continue
			}
		}
	}
}

func (h Hera) handleStartEvent(event events.Message) {
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

	h.registerTunnel(container.ID, tunnel)
}

func (h Hera) handleDieEvent(event events.Message) {
	container, err := NewContainer(h.Client, event.ID)
	if err != nil {
		log.Error(err)
		return
	}

	if tunnel, ok := h.RegisteredTunnels[container.ID]; ok {
		tunnel.stop()
	}
}

func (h Hera) registerTunnel(id string, tunnel *Tunnel) {
	h.RegisteredTunnels[id] = tunnel
}
