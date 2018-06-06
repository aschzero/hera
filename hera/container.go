package main

import (
	"context"
	"fmt"
	"net"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

// Container holds only the container info we care about
type Container struct {
	ID       string
	Hostname string
	Labels   map[string]string
}

// NewContainerFromEvent returns a Container from a given event message
func NewContainerFromEvent(cli *client.Client, event events.Message) (*Container, error) {
	res, err := cli.ContainerInspect(context.Background(), event.ID)
	if err != nil {
		return nil, err
	}

	container := &Container{
		ID:       res.ID,
		Hostname: res.Config.Hostname,
		Labels:   res.Config.Labels,
	}

	return container, nil
}

// VerifyLabelConfig checks the presence of required labels
func (c Container) VerifyLabelConfig() error {
	required := []string{
		"hera.hostname",
		"hera.port",
	}

	for _, label := range required {
		if _, ok := c.Labels[label]; !ok {
			return fmt.Errorf("%s label not found", label)
		}
	}

	return nil
}

// ResolveHostname resolves the container hostname to an address.
func (c Container) ResolveHostname() (string, error) {
	resolved, err := net.LookupHost(c.Hostname)
	if err != nil {
		return "", err
	}

	return resolved[0], nil
}
