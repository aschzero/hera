package main

import (
	"github.com/docker/docker/client"
)

func NewClient() (*client.Client, error) {
	client, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
