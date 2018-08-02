package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

const (
	ServicesPath           = "/etc/services.d"
	RegisteredServicesPath = "/var/run/s6/services"
)

type Tunnel struct {
	ContainerHostname string
	HeraHostname      string
	HeraPort          string
	Certificate       *Certificate
	TunnelConfig      *TunnelConfig
}

type TunnelConfig struct {
	ServicePath           string
	RegisteredServicePath string
	ConfigFilePath        string
	LogFilePath           string
}

func NewTunnelConfig(heraHostname string) *TunnelConfig {
	servicePath := filepath.Join(ServicesPath, heraHostname)
	registeredServicePath := filepath.Join(RegisteredServicesPath, heraHostname)
	configFilePath := filepath.Join(servicePath, "config.yml")
	logFilePath := strings.Join([]string{filepath.Join("/var/log/hera", heraHostname), "log"}, ".")

	tunnelConfig := &TunnelConfig{
		ServicePath:           servicePath,
		RegisteredServicePath: registeredServicePath,
		ConfigFilePath:        configFilePath,
		LogFilePath:           logFilePath,
	}

	return tunnelConfig
}

func NewTunnel(containerHostname string, heraHostname string, heraPort string, certificate *Certificate) *Tunnel {
	tunnelConfig := NewTunnelConfig(heraHostname)

	tunnel := &Tunnel{
		ContainerHostname: containerHostname,
		HeraHostname:      heraHostname,
		HeraPort:          heraPort,
		Certificate:       certificate,
		TunnelConfig:      tunnelConfig,
	}

	return tunnel
}

func (t *Tunnel) start() error {
	log.Infof("Registering tunnel %s @ %s:%s", t.HeraHostname, t.ContainerHostname, t.HeraPort)
	log.Infof("Logging to %s", t.TunnelConfig.LogFilePath)

	if err := t.prepareService(); err != nil {
		return err
	}

	if err := t.generateConfigFile(); err != nil {
		return err
	}

	if err := t.generateRunFile(); err != nil {
		return err
	}

	if err := t.startService(); err != nil {
		return err
	}

	return nil
}

func (t *Tunnel) stop() {
	err := exec.Command("s6-svc", "-d", t.TunnelConfig.ServicePath).Run()
	if err != nil {
		log.Errorf("Unable to stop tunnel %s: %s", t.HeraHostname, err)
		return
	}

	log.Infof("Stopping tunnel %s", t.HeraHostname)
}

func (t *Tunnel) prepareService() error {
	exists, err := afero.DirExists(fs, t.TunnelConfig.ServicePath)
	if err != nil {
		return err
	}

	if !exists {
		fs.Mkdir(t.TunnelConfig.ServicePath, os.ModePerm)
	}

	return nil
}

func (t *Tunnel) generateConfigFile() error {
	configLines := []string{
		"hostname: %s",
		"url: %s:%s",
		"logfile: %s",
		"origincert: %s",
		"no-autoupdate: true",
	}

	config := fmt.Sprintf(strings.Join(configLines[:], "\n"), t.HeraHostname, t.ContainerHostname, t.HeraPort, t.TunnelConfig.LogFilePath, t.Certificate.fullPath())

	if err := afero.WriteFile(fs, t.TunnelConfig.ConfigFilePath, []byte(config), 0644); err != nil {
		return err
	}

	return nil
}

func (t *Tunnel) generateRunFile() error {
	runFilePath := filepath.Join(t.TunnelConfig.ServicePath, "run")
	runLines := []string{
		"#!/bin/sh",
		"exec cloudflared --config %s",
	}

	run := fmt.Sprintf(strings.Join(runLines[:], "\n"), t.TunnelConfig.ConfigFilePath)

	if err := afero.WriteFile(fs, runFilePath, []byte(run), os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (t *Tunnel) startService() error {
	registered, err := t.serviceRegistered()
	if err != nil {
		return err
	}

	if registered {
		running, err := t.serviceRunning()
		if err != nil {
			return err
		}

		if running {
			log.Info("Waiting for previous tunnel to shut down")

			err = exec.Command("s6-svwait", "-d", t.TunnelConfig.RegisteredServicePath).Run()
			if err != nil {
				return err
			}
		}

		err = exec.Command("s6-svc", "-u", t.TunnelConfig.RegisteredServicePath).Run()
		if err != nil {
			return err
		}

		return nil
	}

	err = t.registerService()
	if err != nil {
		return err
	}

	err = exec.Command("s6-svscanctl", "-a", RegisteredServicesPath).Run()
	if err != nil {
		return err
	}

	return nil
}

func (t *Tunnel) registerService() error {
	err := os.Symlink(t.TunnelConfig.ServicePath, t.TunnelConfig.RegisteredServicePath)

	return err
}

func (t *Tunnel) serviceRegistered() (bool, error) {
	registered, err := afero.Exists(fs, t.TunnelConfig.RegisteredServicePath)
	if err != nil {
		return false, err
	}

	return registered, nil
}

func (t *Tunnel) serviceRunning() (bool, error) {
	out, err := exec.Command("s6-svstat", "-u", t.TunnelConfig.RegisteredServicePath).Output()
	if err != nil {
		return false, err
	}

	return strings.Contains(string(out), "true"), nil
}
