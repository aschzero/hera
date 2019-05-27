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

type Handler struct {
	Client *Client
}

func NewHandler(client *Client) *Handler {
	handler := &Handler{
		Client: client,
	}

	return handler
}

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

	log.Infof("Hera container found, connecting to %s...", container.ID[:12])

	ip, err := h.resolveHostname(container)
	if err != nil {
		return err
	}

	cert, err := getCertificate(hostname)
	if err != nil {
		return err
	}

	config := &Config{
		IP:       ip,
		Hostname: hostname,
		Port:     port,
	}

	tunnel := NewTunnel(config, cert)
	tunnel.Start()

	return nil
}

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

func (h *Handler) resolveHostname(container types.ContainerJSON) (string, error) {
	var resolved []string
	var err error
	attempts := 0

	for attempts < 5 {
		attempts++
		resolved, err = net.LookupHost(container.Config.Hostname)

		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		if err == nil {
			return resolved[0], nil
		}
	}

	return "", fmt.Errorf("Unable to connect to %s", container.ID[:12])
}

func getLabel(name string, container types.ContainerJSON) string {
	value, ok := container.Config.Labels[name]
	if !ok {
		return ""
	}

	return value
}

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

func getRootHostname(hostname string) (string, error) {
	httpsHostname := strings.Join([]string{"https://", hostname}, "")

	parsed, err := tld.Parse(httpsHostname)
	if err != nil {
		return "", fmt.Errorf("Unable to parse hostname %s: %s", httpsHostname, err)
	}

	rootHostname := strings.Join([]string{parsed.Domain, parsed.TLD}, ".")

	return rootHostname, nil
}
