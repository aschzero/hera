package main

import (
	"context"
	"errors"
	"fmt"
	"net"
)

type Container struct {
	ID       string
	Hostname string
	Labels   map[string]string
}

func NewContainer(client *Client, id string) (*Container, error) {
	res, err := client.Docker.ContainerInspect(context.Background(), id)
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

func (c Container) tryTunnel() (*Tunnel, error) {
	address, err := c.resolveHostname()
	if err != nil {
		return nil, err
	}

	hostname, err := c.getHostname()
	if err != nil {
		return nil, err
	}

	port, err := c.getPort()
	if err != nil {
		return nil, err
	}

	cert, err := c.getCertificate()
	if err != nil {
		return nil, err
	}

	tunnel := NewTunnel(address, hostname, port, cert)

	return tunnel, nil
}

func (c Container) resolveHostname() (string, error) {
	resolved, err := net.LookupHost(c.Hostname)
	if err != nil {
		return "", fmt.Errorf("unable to resolve hostname %s", c.Hostname)
	}

	return resolved[0], nil
}

func (c Container) getHostname() (string, error) {
	hostname, ok := c.Labels["hera.hostname"]
	if !ok || hostname == "" {
		return "", errors.New("No hera.hostname label")
	}

	return hostname, nil
}

func (c Container) getPort() (string, error) {
	port, ok := c.Labels["hera.port"]
	if !ok || port == "" {
		return "", errors.New("No hera.port label")
	}

	return port, nil
}

func (c Container) getCertificate() (*Certificate, error) {
	name, _ := c.Labels["hera.certificate"]
	cert := NewCertificate(name)

	if cert.isExist() {
		return cert, nil
	}

	hostname, _ := c.getHostname()
	config := NewCertificateConfig()
	certs, err := config.scan()
	if err != nil {
		return nil, fmt.Errorf("Unable to find matching certificate %s: %s", err, cert.fullPath())
	}

	for _, cert := range certs {
		if cert.matchesDomain(hostname) {
			return cert, nil
		}
	}

	return nil, fmt.Errorf("Unable to find matching certificate: %s", cert.fullPath())
}
