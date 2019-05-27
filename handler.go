package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/spf13/afero"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/jpillora/go-tld"
)

const (
	heraHostname = "hera.hostname"
	heraPort     = "hera.port"
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
	switch status := event.Status; status {
	case "start":
		err := h.handleStartEvent(event)
		if err != nil {
			log.Error(err.Error())
		}

	case "die":
		err := h.handleDieEvent(event)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

// HandleContainer allows immediate tunnel creation when hera is started by treating existing
// containers as start events
func (h *Handler) HandleContainer(id string) error {
	event := events.Message{
		ID: id,
	}

	err := h.handleStartEvent(event)
	if err != nil {
		return err
	}

	return nil
}

// handleStartEvent inspects the container from a start event and creates a tunnel if the container
// has been appropriately labeled and a certificate exists for its hostname
func (h *Handler) handleStartEvent(event events.Message) error {
	container, err := h.Client.Inspect(event.ID)
	if err != nil {
		return err
	}

	hostname := getLabel(heraHostname, container)
	port := getLabel(heraPort, container)
	if hostname == "" || port == "" {
		return nil
	}

	log.Infof("Container found, connecting to %s...", container.ID[:12])

	ip, err := h.resolveHostname(container)
	if err != nil {
		return err
	}

	cert, err := getCertificate(hostname)
	if err != nil {
		return err
	}

	config := &TunnelConfig{
		IP:       ip,
		Hostname: hostname,
		Port:     port,
	}

	tunnel := NewTunnel(config, cert)
	tunnel.Start()

	return nil
}

// handleDieEvent inspects the container from a die event and stops the tunnel if one exists.
// An error is returned if a tunnel cannot be found or if the tunnel fails to stop
func (h *Handler) handleDieEvent(event events.Message) error {
	container, err := h.Client.Inspect(event.ID)
	if err != nil {
		return err
	}

	hostname := getLabel("hera.hostname", container)
	if hostname == "" {
		return nil
	}

	tunnel, err := GetTunnelForHost(hostname)
	if err != nil {
		return err
	}

	err = tunnel.Stop()
	if err != nil {
		return err
	}

	return nil
}

// resolveHostname returns the IP address of a container from its hostname.
// An error is returned if the hostname cannot be resolved after five attempts.
func (h *Handler) resolveHostname(container types.ContainerJSON) (string, error) {
	var resolved []string
	var err error

	attempts := 0
	maxAttempts := 5

	for attempts < maxAttempts {
		attempts++
		resolved, err = net.LookupHost(container.Config.Hostname)

		if err != nil {
			time.Sleep(2 * time.Second)
			log.Infof("Unable to connect, retrying... (%d/%d)", attempts, maxAttempts)

			continue
		}

		return resolved[0], nil
	}

	return "", fmt.Errorf("Unable to connect to %s", container.ID[:12])
}

// getLabel returns the label value from a given label name and container JSON.
func getLabel(name string, container types.ContainerJSON) string {
	value, ok := container.Config.Labels[name]
	if !ok {
		return ""
	}

	return value
}

// getCertificate returns a Certificate for a given hostname.
// An error is returned if the root hostname cannot be parsed or if the certificate cannot be found.
func getCertificate(hostname string) (*Certificate, error) {
	rootHostname, err := getRootHostname(hostname)
	if err != nil {
		return nil, err
	}

	cert, err := FindCertificateForHost(rootHostname, afero.NewOsFs())
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// getRootHostname parses and validates a URL and returns the root hostname (e.g.: domain.tld).
// An error is returned if the hostname does not contain a valid TLD.
func getRootHostname(hostname string) (string, error) {
	httpsHostname := strings.Join([]string{"https://", hostname}, "")

	parsed, err := tld.Parse(httpsHostname)
	if err != nil {
		return "", fmt.Errorf("Unable to parse hostname %s: %s", httpsHostname, err)
	}

	rootHostname := strings.Join([]string{parsed.Domain, parsed.TLD}, ".")

	return rootHostname, nil
}
