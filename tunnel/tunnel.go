package tunnel

import (
	"fmt"
	"os"
	"strings"

	"github.com/aschzero/hera/certificate"
	logging "github.com/op/go-logging"
	"github.com/spf13/afero"
)

var (
	registry = make(map[string]*Tunnel)
	log      = logging.MustGetLogger("hera")
)

type Tunnel struct {
	Config      *Config
	Certificate *certificate.Certificate
	Service     *Service
}

type Config struct {
	IP       string
	Hostname string
	Port     string
}

func New(config *Config, certificate *certificate.Certificate) *Tunnel {
	service := NewService(config.Hostname)

	tunnel := &Tunnel{
		Config:      config,
		Certificate: certificate,
		Service:     service,
	}

	return tunnel
}

func Get(hostname string) (*Tunnel, error) {
	tunnel, ok := registry[hostname]

	if !ok {
		return nil, fmt.Errorf("No tunnel exists for %s", hostname)
	}

	return tunnel, nil
}

func (t *Tunnel) Start() error {
	err := t.prepareService()
	if err != nil {
		return err
	}

	err = t.startService()
	if err != nil {
		return err
	}

	registry[t.Config.Hostname] = t

	return nil
}

func (t *Tunnel) Stop() error {
	log.Infof("Stopping tunnel %s", t.Config.Hostname)

	err := t.Service.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (t *Tunnel) prepareService() error {
	err := t.Service.Create()
	if err != nil {
		return err
	}

	err = t.writeConfigFile()
	if err != nil {
		return err
	}

	err = t.writeRunFile()
	if err != nil {
		return err
	}

	return nil
}

func (t *Tunnel) startService() error {
	supervised, err := t.Service.IsSupervised()
	if err != nil {
		return err
	}

	if !supervised {
		log.Infof("Registering tunnel %s", t.Config.Hostname)

		err := t.Service.Supervise()
		if err != nil {
			return err
		}
		return nil
	}

	running, err := t.Service.IsRunning()
	if err != nil {
		return err
	}

	if running {
		log.Infof("Restarting tunnel %s", t.Config.Hostname)

		err := t.Service.Restart()
		if err != nil {
			return err
		}
	} else {
		log.Infof("Starting tunnel %s", t.Config.Hostname)

		err := t.Service.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Tunnel) writeConfigFile() error {
	configLines := []string{
		"hostname: %s",
		"url: %s:%s",
		"logfile: %s",
		"origincert: %s",
		"no-autoupdate: true",
	}

	contents := fmt.Sprintf(strings.Join(configLines[:], "\n"), t.Config.Hostname, t.Config.IP, t.Config.Port, t.Service.LogFilePath(), t.Certificate.FullPath())

	err := afero.WriteFile(fs, t.Service.ConfigFilePath(), []byte(contents), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tunnel) writeRunFile() error {
	runLines := []string{
		"#!/bin/sh",
		"exec cloudflared --config %s",
	}
	contents := fmt.Sprintf(strings.Join(runLines[:], "\n"), t.Service.ConfigFilePath())

	err := afero.WriteFile(fs, t.Service.RunFilePath(), []byte(contents), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
