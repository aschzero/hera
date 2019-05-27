package main

import (
    "io"

    "github.com/spf13/afero"
)

// Listener holds config for an event listener and is used to listen for container events
type Listener struct {
	Client *Client
	Fs     afero.Fs
}

// NewListener returns a new Listener
func NewListener() (*Listener, error) {
	client, err := NewClient()
	if err != nil {
		log.Errorf("Unable to connect to Docker: %s", err)
		return nil, err
	}

	listener := &Listener{
		Client: client,
		Fs:     afero.NewOsFs(),
	}

	return listener, nil
}

// Revive revives tunnels for currently running containers
func (l *Listener) Revive() error {
	handler := NewHandler(l.Client)
	containers, err := l.Client.ListContainers()
	if err != nil {
		return err
	}

	for _, c := range containers {
		err := handler.HandleContainer(c.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Listen listens for container events to be handled
func (l *Listener) Listen() {
	log.Info("Hera is listening")

	handler := NewHandler(l.Client)
	messages, errs := l.Client.Events()

	for {
		select {
		case event := <-messages:
			handler.HandleEvent(event)

		case err := <-errs:
			if err != nil && err != io.EOF {
				log.Error(err.Error())
			}
		}
	}
}
