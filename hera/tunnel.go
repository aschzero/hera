package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/afero"
)

// Tunnel holds tunnel metadata and configuration
type Tunnel struct {
	ContainerHostname string
	HeraHostname      string
	HeraPort          string
	TunnelConfig      *TunnelConfig
}

// TunnelConfig holds tunnel configuration
type TunnelConfig struct {
	ServicePath         string
	RunFilePath         string
	S6ServicesPath      string
	S6TunnelServicePath string
	ConfigFilePath      string
	LogFilePath         string
}

// NewTunnel returns a Tunnel with the necessary metadata and configuration
func NewTunnel(containerHostname string, heraHostname string, heraPort string) *Tunnel {
	servicePath := filepath.Join("/etc/services.d", heraHostname)
	runFilePath := filepath.Join(servicePath, "run")
	s6ServicesPath := "/var/run/s6/services"
	s6TunnelServicePath := filepath.Join(s6ServicesPath, heraHostname)
	configFilePath := filepath.Join(servicePath, "config.yml")
	logFilePath := filepath.Join("/var/log/hera", heraHostname+".log")

	tunnelConfig := &TunnelConfig{
		ServicePath:         servicePath,
		RunFilePath:         runFilePath,
		S6ServicesPath:      s6ServicesPath,
		S6TunnelServicePath: s6TunnelServicePath,
		ConfigFilePath:      configFilePath,
		LogFilePath:         logFilePath,
	}

	tunnel := &Tunnel{
		ContainerHostname: containerHostname,
		HeraHostname:      heraHostname,
		HeraPort:          heraPort,
		TunnelConfig:      tunnelConfig,
	}

	return tunnel
}

// Start starts a tunnel
func (t Tunnel) Start() error {
	log.Infof("\nRegistering tunnel %s @ %s:%s", t.HeraHostname, t.ContainerHostname, t.HeraPort)
	log.Infof("Logging to %s\n\n", t.TunnelConfig.LogFilePath)

	if err := t.PrepareService(); err != nil {
		return err
	}

	if err := t.GenerateConfigFile(); err != nil {
		return err
	}

	if err := t.GenerateRunFile(); err != nil {
		return err
	}

	if err := t.StartService(); err != nil {
		return err
	}

	return nil
}

// Stop stops a tunnel
func (t Tunnel) Stop() {
	if err := exec.Command("s6-svc", []string{"-d", t.TunnelConfig.ServicePath}...).Run(); err != nil {
		log.Errorf("Error while stopping tunnel %s: %s", t.HeraHostname, err)
		return
	}

	log.Infof("\nStopped tunnel %s\n\n", t.HeraHostname)
}

// PrepareService creates the tunnel service directory if it doesn't exist
func (t Tunnel) PrepareService() error {
	exists, err := afero.DirExists(fs, t.TunnelConfig.ServicePath)
	if err != nil {
		return err
	}

	if !exists {
		fs.Mkdir(t.TunnelConfig.ServicePath, os.ModePerm)
	}

	return nil
}

// GenerateConfigFile generates a new cloudflared config file
func (t Tunnel) GenerateConfigFile() error {
	config := fmt.Sprintf("hostname: %s\nurl: %s:%s\nlogfile: %s", t.HeraHostname, t.ContainerHostname, t.HeraPort, t.TunnelConfig.LogFilePath)

	if err := afero.WriteFile(fs, t.TunnelConfig.ConfigFilePath, []byte(config), 0644); err != nil {
		return err
	}

	return nil
}

// GenerateRunFile generates the tunnel service run file with
// the necessary permissions.
func (t Tunnel) GenerateRunFile() error {
	run := fmt.Sprintf("#!/usr/bin/with-contenv sh\nexec cloudflared --config %s", t.TunnelConfig.ConfigFilePath)

	if err := afero.WriteFile(fs, t.TunnelConfig.RunFilePath, []byte(run), os.ModePerm); err != nil {
		return err
	}

	return nil
}

// StartService starts the tunnel service
func (t Tunnel) StartService() error {
	exists, err := afero.Exists(fs, t.TunnelConfig.S6TunnelServicePath)
	if err != nil {
		return err
	}

	if exists {
		if err := exec.Command("s6-svc", []string{"-u", t.TunnelConfig.S6TunnelServicePath}...).Run(); err != nil {
			return err
		}
	} else {
		if err := os.Symlink(t.TunnelConfig.ServicePath, t.TunnelConfig.S6TunnelServicePath); err != nil {
			return err
		}

		if err := exec.Command("s6-svscanctl", []string{"-a", t.TunnelConfig.S6ServicesPath}...).Run(); err != nil {
			return err
		}
	}

	return nil
}
