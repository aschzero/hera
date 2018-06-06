package main

import (
	"context"
	"fmt"
	"io"
	"net"
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
	container, err := h.Client.ContainerInspect(context.Background(), event.ID)
	if err != nil {
		log.Error(err)
	}

	labels := container.Config.Labels
	if err := verifyLabelConfig(labels); err != nil {
		log.Errorf("Ignoring container %s: %s", container.ID, err)
		return
	}

	heraHostname, _ := labels["hera.hostname"]
	heraPort, _ := labels["hera.port"]

	hostname := container.Config.Hostname
	resolved, err := resolvedHostname(hostname)
	if err != nil {
		log.Errorf("Unable to resolve hostname %s for container %s. Ensure the container is accessible within Hera's network.", hostname, container.ID)
		return
	}

	tunnel := NewTunnel(resolved, heraHostname, heraPort)
	h.ActiveTunnels[hostname] = tunnel

	err = tunnel.Start()
	if err != nil {
		log.Errorf("Error trying to start tunnel: %s", err)
	}
}

// HandleDieEvent stops a tunnel if it exists when a container is stopped.
func (h Hera) HandleDieEvent(event events.Message) {
	container, err := h.Client.ContainerInspect(context.Background(), event.ID)
	if err != nil {
		log.Errorf("Error trying to stop tunnel: %s", err)
		return
	}

	hostname := container.Config.Hostname
	if tunnel, ok := h.ActiveTunnels[hostname]; ok {
		tunnel.Stop()
	}
}

func verifyLabelConfig(labels map[string]string) error {
	required := []string{
		"hera.hostname",
		"hera.port",
	}

	for _, label := range required {
		if _, ok := labels[label]; !ok {
			return fmt.Errorf("%s label not found", label)
		}
	}

	return nil
}

func resolvedHostname(hostname string) (string, error) {
	resolved, err := net.LookupHost(hostname)
	if err != nil {
		return "", err
	}

	return resolved[0], nil
}
