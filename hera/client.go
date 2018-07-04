package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

type DockerClient interface {
	ContainerInspect(context.Context, string) (types.ContainerJSON, error)
	ContainerList(context.Context, types.ContainerListOptions) ([]types.Container, error)
	Events(context.Context, types.EventsOptions) (<-chan events.Message, <-chan error)
}

type Client struct {
	Docker DockerClient
}

func NewClient() (*Client, error) {
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Docker: cli,
	}

	return client, nil
}
