package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

const (
	Socket     = "unix:///var/run/docker.sock"
	APIVersion = "v1.22"
)

// Client holds an instance of the docker client
type Client struct {
	DockerClient *client.Client
}

// NewClient returns a new Client or an error if not able to connect to the Docker daemon
func NewClient() (*Client, error) {
	cli, err := client.NewClient(Socket, APIVersion, nil, nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		DockerClient: cli,
	}

	return client, nil
}

// Events returns a channel of Docker events
func (c *Client) Events() (<-chan events.Message, <-chan error) {
	return c.DockerClient.Events(context.Background(), types.EventsOptions{})
}

// ListContainers returns a collection of Docker containers
func (c *Client) ListContainers() ([]types.Container, error) {
	return c.DockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
}

// ListServices returns a collection of Docker services
func (c *Client) ListServices() ([]swarm.Service, error) {
	return c.DockerClient.ServiceList(context.Background(), types.ServiceListOptions{})
}

// Inspect returns the full information for a container with the given container ID
func (c *Client) Inspect(id string) (types.ContainerJSON, error) {
	return c.DockerClient.ContainerInspect(context.Background(), id)
}

// InspectSvc returns the full information for a container with the given container ID
func (c *Client) InspectSvc(id string) (swarm.Service, []byte, error) {
	return c.DockerClient.ServiceInspectWithRaw(context.Background(), id, types.ServiceInspectOptions{})
}

// FindNetwork returns the full information for a container with the given container ID
func (c *Client) InspectNetwork(id string) (types.NetworkResource, error) {
	return c.DockerClient.NetworkInspect(context.Background(), id, types.NetworkInspectOptions{})
}
