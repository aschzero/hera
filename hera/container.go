package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	tld "github.com/jpillora/go-tld"
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
		ID:       res.ID[0:12],
		Hostname: res.Config.Hostname,
		Labels:   res.Config.Labels,
	}

	return container, nil
}

func (c *Container) tryTunnel() (*Tunnel, error) {
	ip, err := c.resolveHostname()
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

	tunnel := NewTunnel(ip, hostname, c.Hostname, port, cert)

	return tunnel, nil
}

func (c *Container) hasRequiredLabels() bool {
	required := []string{
		"hera.hostname",
		"hera.port",
	}

	for _, label := range required {
		if _, ok := c.Labels[label]; !ok {
			return false
		}
	}

	return true
}

func (c *Container) resolveHostname() (string, error) {
	var resolved []string
	var err error
	attempts := 0
	maxAttempts := 5

	for attempts < maxAttempts {
		attempts++
		resolved, err = net.LookupHost(c.Hostname)

		if err != nil {
			log.Infof("Unable to resolve hostname, retrying (%d/%d)", attempts, maxAttempts)
			time.Sleep(2 * time.Second)
			continue
		}

		if err == nil {
			return resolved[0], nil
		}
	}

	return "", fmt.Errorf("Unable to resolve hostname for container %s", c.ID)
}

func (c *Container) getHostname() (string, error) {
	hostname, ok := c.Labels["hera.hostname"]
	if !ok || hostname == "" {
		return "", errors.New("No hera.hostname label")
	}

	return hostname, nil
}

func (c *Container) getPort() (string, error) {
	port, ok := c.Labels["hera.port"]
	if !ok || port == "" {
		return "", errors.New("No hera.port label")
	}

	return port, nil
}

func (c *Container) getCertificate() (*Certificate, error) {
	certConfig := NewCertificateConfig()
	rootHostname, err := c.getRootHost()
	if err != nil {
		return nil, err
	}

	cert, err := certConfig.findCertificateForHost(rootHostname)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *Container) getRootHost() (string, error) {
	hostname, err := c.getHostname()
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(hostname, "http") {
		hostname = "https://" + hostname
	}

	parsed, err := tld.Parse(hostname)
	if err != nil {
		return "", err
	}

	root := parsed.Domain + "." + parsed.TLD
	return root, nil
}
