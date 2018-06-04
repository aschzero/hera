package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("hera")

type Hera struct {
	Client        *client.Client
	ActiveTunnels map[string]*Tunnel
}

func main() {
	LogInit()

	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, nil)
	if err != nil {
		log.Errorf("Error when trying to connect to the Docker daemon: %s", err)
		return
	}

	hera := &Hera{
		Client:        cli,
		ActiveTunnels: make(map[string]*Tunnel),
	}

	certificate := NewCertificate()
	certificate.VerifyCertificate()

	hera.Listen()
}

func LogInit() {
	logFile, err := os.OpenFile("/var/log/hera/hera.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Sprintf("Unable open log file: %s", err))
	}

	stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logFileBackend := logging.NewLogBackend(logFile, "", 0)
	logFileBackendFormat := logging.MustStringFormatter(
		`%{time:15:04:00.000} [%{level}] %{message}`,
	)
	logFileBackendFormatter := logging.NewBackendFormatter(logFileBackend, logFileBackendFormat)

	logging.SetBackend(stderrBackend, logFileBackendFormatter)
}

func (h Hera) Listen() {
	log.Info("\nHera is listening...\n")

	messages, errs := h.Client.Events(context.Background(), types.EventsOptions{})

	for {
		select {
		case err := <-errs:
			if err != nil && err != io.EOF {
				log.Error(err)
			}

			os.Exit(1)

		case event := <-messages:
			if event.Status == "start" {
				h.HandleStartEvent(event)
				continue
			}

			if event.Status == "die" {
				h.HandleDieEvent(event)
				continue
			}
		}
	}
}

func (h Hera) HandleStartEvent(event events.Message) {
	container, err := h.Client.ContainerInspect(context.Background(), event.ID)
	if err != nil {
		log.Error(err)
	}

	labels := container.Config.Labels
	heraHostname, heraHostnamePresent := labels["hera.hostname"]
	heraPort, heraPortPresent := labels["hera.port"]
	if !heraHostnamePresent || !heraPortPresent {
		log.Infof("Ignoring container %s: no hera labels found", event.ID)
		return
	}

	hostname := container.Config.Hostname
	resolved, err := net.LookupHost(hostname)
	if err != nil {
		log.Errorf("Unable to resolve hostname %s for container %s. Ensure the container is accessible within Hera's network.", hostname, container.ID)
		return
	}

	tunnel := NewTunnel(resolved[0], heraHostname, heraPort)
	h.ActiveTunnels[hostname] = tunnel

	err = tunnel.Start()
	if err != nil {
		log.Errorf("Error while trying to start tunnel: %s", err)
	}
}

func (h Hera) HandleDieEvent(event events.Message) {
	container, err := h.Client.ContainerInspect(context.Background(), event.ID)
	if err != nil {
		log.Errorf("Error while trying to stop tunnel: %s", err)
		return
	}

	hostname := container.Config.Hostname
	if tunnel, tunnelPresent := h.ActiveTunnels[hostname]; tunnelPresent {
		tunnel.Stop()
	}
}
