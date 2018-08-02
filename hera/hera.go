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

func (h *Hera) listen() {
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

func (h *Hera) handleStartEvent(event events.Message) {
	err := h.tryTunnel(event.ID, true)
	if err != nil {
		log.Error(err)
	}
}

func (h *Hera) handleDieEvent(event events.Message) {
	container, err := NewContainer(h.Client, event.ID)
	if err != nil {
		log.Error(err)
		return
	}

	if tunnel, ok := h.RegisteredTunnels[container.ID]; ok {
		tunnel.stop()
	}
}

func (h *Hera) revive() {
	containers, err := h.Client.Docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Error(err)
		return
	}

	for _, c := range containers {
		err := h.tryTunnel(c.ID, false)
		if err != nil {
			log.Error(err)
		}
	}
}

func (h *Hera) tryTunnel(id string, logIgnore bool) error {
	container, err := NewContainer(h.Client, id)
	if err != nil {
		return err
	}

	hasLabels := container.hasRequiredLabels()
	if !hasLabels {
		if logIgnore {
			log.Infof("Ignoring container %s", container.ID)
		}
		return nil
	}

	tunnel, err := container.tryTunnel()
	if err != nil {
		return err
	}

	if err := tunnel.start(); err != nil {
		return err
	}

	h.registerTunnel(container.ID, tunnel)
	return nil
}

func (h *Hera) registerTunnel(id string, tunnel *Tunnel) {
	h.RegisteredTunnels[id] = tunnel
}
