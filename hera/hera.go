package main

import (
	"io"

	"github.com/docker/docker/api/types/events"
)

type Hera struct {
	Client            *Client
	RegisteredTunnels map[string]*Tunnel
}

func run() {
	client, err := NewClient()
	if err != nil {
		log.Errorf("Unable to connect to Docker: %s", err)
		return
	}

	hera := &Hera{
		Client:            client,
		RegisteredTunnels: make(map[string]*Tunnel),
	}

	err = hera.revive()
	if err != nil {
		log.Error(err)
	}

	hera.listen()
}

func (h *Hera) listen() {
	log.Info("Hera is listening")

	messages, errs := h.Client.events()

	for {
		select {
		case err := <-errs:
			if err != nil && err != io.EOF {
				log.Error(err)
			}

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
	container, err := NewContainer(event.ID, h.Client)
	if err != nil {
		log.Error(err)
		return
	}

	if tunnel, ok := h.RegisteredTunnels[container.ID]; ok {
		err := tunnel.stop()
		if err != nil {
			log.Errorf("Unable to stop tunnel %s: %s", tunnel.Config.TunnelHostname, err)
		}
	}
}

func (h *Hera) revive() error {
	containers, err := h.Client.listContainers()
	if err != nil {
		return err
	}

	for _, c := range containers {
		err := h.tryTunnel(c.ID, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Hera) tryTunnel(id string, logIgnore bool) error {
	container, err := NewContainer(id, h.Client)
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

	err = tunnel.start()
	if err != nil {
		return err
	}

	h.registerTunnel(container.ID, tunnel)
	return nil
}

func (h *Hera) registerTunnel(id string, tunnel *Tunnel) {
	h.RegisteredTunnels[id] = tunnel
}
