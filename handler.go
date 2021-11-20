package main

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/spf13/afero"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/swarm"
)

// A Handler is responsible for responding to container start and die events
type Handler struct {
	Client *Client
}

// NewHandler returns a new Handler instance
func NewHandler(client *Client) *Handler {
	handler := &Handler{
		Client: client,
	}

	return handler
}

// HandleEvent dispatches an event to the appropriate handler method depending on its status
func (h *Handler) HandleEvent(event events.Message) {

	switch status := event.Action; status {
	case "start", "update":
		err := h.handleStartEvent(event)
		if err != nil {
			log.Error(err.Error())
		}

	case "die", "remove":
		err := h.handleDieEvent(event)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

// handleStartEvent inspects the container from a start event and creates a tunnel if the container
// has been appropriately labeled and a certificate exists for its hostname
func (h *Handler) handleStartEvent(event events.Message) error {

	var (
		hostname string
		port     string
		ip       string
		id       string
	)

	if swarmMode {
		if event.Type != events.ServiceEventType {
			return nil
		}

		id = event.Actor.ID
		service, _, err := h.Client.InspectService(id)
		if err != nil {
			return err
		}

		hostname = getServiceLabel(heraHostLabel, service)
		port = getServiceLabel(heraPortLabel, service)
		if hostname == "" || port == "" {
			return nil
		}

		log.Infof("Service found, connecting to %s...", service.ID[:12])

		ip, err = h.resolveHostname(event.Actor.Attributes["name"])
		if err != nil {
			return err
		}

	} else {

		id = event.ID
		container, err := h.Client.InspectContainer(id)
		if err != nil {
			return err
		}

		hostname = getLabel(heraHostLabel, container)
		port = getLabel(heraPortLabel, container)
		if hostname == "" || port == "" {
			return nil
		}

		log.Infof("Container found, connecting to %s...", container.ID[:12])

		ip, err = h.resolveHostname(container.Config.Hostname)
		if err != nil {
			return err
		}
	}

	cert, err := getCertificate(hostname)
	if err != nil {
		return err
	}

	config := &TunnelConfig{
		ID:       id,
		IP:       ip,
		Hostname: hostname,
		Port:     port,
	}

	tunnel := NewTunnel(config, cert)
	return tunnel.Start()
}

// handleDieEvent inspects the container from a die event and stops the tunnel if one exists.
// An error is returned if a tunnel cannot be found or if the tunnel fails to stop
func (h *Handler) handleDieEvent(event events.Message) error {

	var id string

	if swarmMode {
		id = event.Actor.ID
	} else {
		id = event.ID
	}

	tunnel, hadTunnel := registry[id]
	if !hadTunnel {
		return nil
	}

	return tunnel.Stop()
}

// resolveHostname returns the IP address of a container from its hostname.
// An error is returned if the hostname cannot be resolved after five attempts.
func (h *Handler) resolveHostname(hostname string) (string, error) {
	var resolved []string
	var err error

	attempts := 0
	maxAttempts := 5

	for attempts < maxAttempts {
		attempts++
		resolved, err = net.LookupHost(hostname)

		if err != nil {
			time.Sleep(2 * time.Second)
			log.Infof("Unable to connect, retrying... (%d/%d)", attempts, maxAttempts)

			continue
		}

		return resolved[0], nil
	}

	return "", fmt.Errorf("Unable to connect to %s", hostname)
}

// getLabel returns the label value from a given label name and container JSON.
func getLabel(name string, container types.ContainerJSON) string {
	value, ok := container.Config.Labels[name]
	if !ok {
		return ""
	}

	return value
}

// getServiceLabel returns the label value from a given label name and service
func getServiceLabel(name string, service swarm.Service) string {
	value, ok := service.Spec.Labels[name]
	if !ok {
		return ""
	}

	return value
}

// getCertificate returns a Certificate for a given hostname.
// An error is returned if the root hostname cannot be parsed or if the certificate cannot be found.
func getCertificate(hostname string) (*Certificate, error) {
	rootHostname, err := getRootDomain(hostname)
	if err != nil {
		return nil, err
	}

	cert, err := FindCertificateForHost(rootHostname, afero.NewOsFs())
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// getRootDomain returns the root domain for a given hostname
func getRootDomain(hostname string) (string, error) {
	domain, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil {
		return "", err
	}

	return domain, nil
}
