package listener

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

const (
	Socket     = "unix:///var/run/docker.sock"
	APIVersion = "v1.22"
)

type Client struct {
	DockerClient *client.Client
}

func newClient() (*Client, error) {
	cli, err := client.NewClient(Socket, APIVersion, nil, nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		DockerClient: cli,
	}

	return client, nil
}

func (c *Client) Events() (<-chan events.Message, <-chan error) {
	return c.DockerClient.Events(context.Background(), types.EventsOptions{})
}

func (c *Client) ListContainers() ([]types.Container, error) {
	return c.DockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
}

func (c *Client) Inspect(id string) (types.ContainerJSON, error) {
	return c.DockerClient.ContainerInspect(context.Background(), id)
}
