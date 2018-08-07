package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/afero"
)

type Tunnel struct {
	Config      *TunnelConfig
	Certificate *Certificate
	Service     *Service
}

type TunnelConfig struct {
	IP             string
	Hostname       string
	TunnelHostname string
	TunnelPort     string
}

func NewTunnel(config *TunnelConfig, certificate *Certificate) *Tunnel {
	serviceConfig := &ServiceConfig{
		Hostname:       config.Hostname,
		TunnelHostname: config.TunnelHostname,
	}
	service := NewService(serviceConfig)

	tunnel := &Tunnel{
		Config:      config,
		Certificate: certificate,
		Service:     service,
	}

	return tunnel
}

func (t *Tunnel) start() error {
	log.Infof("Registering tunnel %s @ %s:%s", t.Config.TunnelHostname, t.Config.IP, t.Config.TunnelPort)

	err := t.prepareService()
	if err != nil {
		return err
	}

	err = t.startService()
	if err != nil {
		return err
	}

	return nil
}

func (t *Tunnel) stop() error {
	err := t.Service.stop()
	if err != nil {
		return err
	}

	log.Infof("Stopping tunnel %s", t.Config.TunnelHostname)
	return nil
}

func (t *Tunnel) prepareService() error {
	err := t.Service.create()
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
	supervised, err := t.Service.isSupervised()
	if err != nil {
		return err
	}

	if !supervised {
		err := t.Service.supervise()
		if err != nil {
			return err
		}
		return nil
	}

	running, err := t.Service.isRunning()
	if err != nil {
		return err
	}

	if running {
		err := t.Service.restart()
		if err != nil {
			return err
		}
	} else {
		err := t.Service.start()
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

	contents := fmt.Sprintf(strings.Join(configLines[:], "\n"), t.Config.TunnelHostname, t.Config.IP, t.Config.TunnelPort, t.Service.logFilePath(), t.Certificate.fullPath())

	err := afero.WriteFile(fs, t.Service.configFilePath(), []byte(contents), 0644)
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
	contents := fmt.Sprintf(strings.Join(runLines[:], "\n"), t.Service.configFilePath())

	err := afero.WriteFile(fs, t.Service.runFilePath(), []byte(contents), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
