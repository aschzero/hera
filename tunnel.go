package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/afero"
)

var (
	registry = make(map[string]*Tunnel)
)

// Tunnel holds the corresponding config, certificate, and service for a tunnel
type Tunnel struct {
	Config      *TunnelConfig
	Certificate *Certificate
	Service     *Service
}

// TunnelConfig holds the necessary configuration for a tunnel
type TunnelConfig struct {
	ID       string
	IP       string
	Hostname string
	Port     string
}

// NewTunnel returns a Tunnel with its corresponding config and certificate
func NewTunnel(config *TunnelConfig, certificate *Certificate) *Tunnel {
	service := NewService(config.Hostname)

	tunnel := &Tunnel{
		Config:      config,
		Certificate: certificate,
		Service:     service,
	}

	return tunnel
}

// Start starts a tunnel
func (t *Tunnel) Start() error {
	err := t.prepareService()
	if err != nil {
		return err
	}

	err = t.startService()
	if err != nil {
		return err
	}

	registry[t.Config.ID] = t

	return nil
}

// Stop stops a tunnel
func (t *Tunnel) Stop() error {
	log.Infof("Stopping tunnel %s", t.Config.Hostname)

	err := t.Service.Stop()
	if err != nil {
		return err
	}
	registry[t.Config.ID] = nil
	return nil
}

// prepareService creates the service and necessary files for the tunnel service
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

// startService starts the tunnel service
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

// writeConfigFile creates the config file for a tunnel
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

// writeRunFile creates the run file for a tunnel
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
