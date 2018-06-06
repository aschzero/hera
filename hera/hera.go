package main

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

// Hera holds an instantiated Client and a map of active tunnels
type Hera struct {
	Client        *client.Client
	ActiveTunnels map[string]*Tunnel
}

// Listen continuously listens for container start or die events.
func (h Hera) Listen() {
	log.Info("Hera is listening...\n\n")

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

// HandleStartEvent checks for valid Hera configuration and creates a new
// tunnel whenn applicable.
func (h Hera) HandleStartEvent(event events.Message) {
	container, err := NewContainerFromEvent(h.Client, event)
	if err != nil {
		log.Error(err)
		return
	}

	if err := container.VerifyLabelConfig(); err != nil {
		log.Errorf("Ignoring container %s: %s", container.ID, err)
		return
	}

	hostname, err := container.ResolveHostname()
	if err != nil {
		log.Errorf("Unable to resolve hostname %s for container %s. Ensure the container is accessible within Hera's network.", container.Hostname, container.ID)
		return
	}

	heraHostname, _ := container.Labels["hera.hostname"]
	heraPort, _ := container.Labels["hera.port"]
	tunnel := NewTunnel(hostname, heraHostname, heraPort)
	h.ActiveTunnels[container.ID] = tunnel

	err = tunnel.Start()
	if err != nil {
		log.Errorf("Error starting tunnel: %s", err)
	}
}

// HandleDieEvent stops a tunnel if it exists when a container is stopped.
func (h Hera) HandleDieEvent(event events.Message) {
	container, err := NewContainerFromEvent(h.Client, event)
	if err != nil {
		log.Error(err)
		return
	}

	if tunnel, ok := h.ActiveTunnels[container.ID]; ok {
		tunnel.Stop()
	}
}
