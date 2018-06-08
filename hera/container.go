package main

import (
	"context"
	"fmt"
	"net"

	"github.com/docker/docker/client"
)

// Container holds only the container info we care about
type Container struct {
	ID       string
	Hostname string
	Labels   map[string]string
}

// NewContainer returns a new Container from a given container ID
func NewContainer(cli *client.Client, id string) (*Container, error) {
	res, err := cli.ContainerInspect(context.Background(), id)
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

// TryTunnel returns a Tunnel if the container is correctly configured
func (c Container) TryTunnel() (*Tunnel, error) {
	if err := c.VerifyLabels(); err != nil {
		return nil, err
	}

	hostname, err := c.ResolveHostname()
	if err != nil {
		return nil, err
	}

	heraHostname, _ := c.Labels["hera.hostname"]
	heraPort, _ := c.Labels["hera.port"]
	tunnel := NewTunnel(hostname, heraHostname, heraPort)

	return tunnel, nil
}

// VerifyLabels checks the presence of required labels
func (c Container) VerifyLabels() error {
	required := []string{
		"hera.hostname",
		"hera.port",
	}

	for _, label := range required {
		if _, ok := c.Labels[label]; !ok {
			return fmt.Errorf("missing labels")
		}
	}

	return nil
}

// ResolveHostname resolves the container hostname to an address.
func (c Container) ResolveHostname() (string, error) {
	resolved, err := net.LookupHost(c.Hostname)
	if err != nil {
		return "", fmt.Errorf("unable to resolve hostname %s", resolved)
	}

	return resolved[0], nil
}
