package main

import (
    "io"

    "github.com/spf13/afero"
)

type Listener struct {
	Client *Client
	Fs     afero.Fs
}

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
