package main

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

// Hera holds an instantiated Client and a map of registered tunnels
type Hera struct {
	Client            *client.Client
	RegisteredTunnels map[string]*Tunnel
}

// Revive starts tunnels for containers already running
func (h Hera) Revive() {
	containers, err := h.Client.ContainerList(context.Background(), types.ContainerListOptions{})
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

		tunnel, err := container.TryTunnel()
		if err != nil {
			continue
		}

		if err := tunnel.Start(); err != nil {
			log.Errorf("Error starting tunnel: %s", err)
			continue
		}

		h.RegisterTunnel(container.ID, tunnel)
	}
}

// Listen continuously listens for container start or die events
func (h Hera) Listen() {
	log.Info("Hera is listening")

	messages, errs := h.Client.Events(context.Background(), types.EventsOptions{})

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

	tunnel, err := container.TryTunnel()
	if err != nil {
		log.Errorf("Ignoring container %s: %s", container.ID, err)
		return
	}

	if err := tunnel.Start(); err != nil {
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
		tunnel.Stop()
	}
}

// RegisterTunnel stores a Tunnel in memory for later reference
func (h Hera) RegisterTunnel(id string, tunnel *Tunnel) {
	h.RegisteredTunnels[id] = tunnel
}
