package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	err := t.PrepareService()
	if err != nil {
		log.Errorf("Error while preparing service for tunnel: %s", err)
		return err
	}

	err = t.GenerateConfigFile()
	if err != nil {
		log.Errorf("Error while generating config file for tunnel: %s", err)
		return err
	}

	err = t.GenerateRunFile()
	if err != nil {
		log.Errorf("Error while generating run file for tunnel: %s", err)
		return err
	}

	err = t.StartService()
	if err != nil {
		log.Errorf("Error while trying to start service for tunnel: %s", err)
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
	if _, err := os.Stat(t.TunnelConfig.ServicePath); os.IsNotExist(err) {
		err := os.Mkdir(t.TunnelConfig.ServicePath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// GenerateConfigFile generates a new cloudflared config file
func (t Tunnel) GenerateConfigFile() error {
	configFile, err := os.Create(t.TunnelConfig.ConfigFilePath)

	defer configFile.Close()

	lines := []string{
		fmt.Sprintf("hostname: %s", t.HeraHostname),
		fmt.Sprintf("url: %s:%s", t.ContainerHostname, t.HeraPort),
		fmt.Sprintf("logfile: %s", t.TunnelConfig.LogFilePath),
	}

	writer := bufio.NewWriter(configFile)
	for _, line := range lines {
		writer.WriteString(line + "\n")
	}

	writer.Flush()

	return err
}

// GenerateRunFile generates the tunnel service run file with
// the necessary permissions.
func (t Tunnel) GenerateRunFile() error {
	runFile, err := os.Create(t.TunnelConfig.RunFilePath)
	if err != nil {
		return err
	}

	defer runFile.Close()

	lines := []string{
		"#!/usr/bin/with-contenv sh\n",
		"exec cloudflared --config " + t.TunnelConfig.ConfigFilePath,
	}

	writer := bufio.NewWriter(runFile)
	for _, line := range lines {
		writer.WriteString(line)
	}

	writer.Flush()

	err = os.Chmod(t.TunnelConfig.RunFilePath, 0755)
	if err != nil {
		return err
	}

	return nil
}

// StartService starts the tunnel service
func (t Tunnel) StartService() error {
	if _, err := os.Stat(t.TunnelConfig.S6TunnelServicePath); err == nil {
		if err := exec.Command("s6-svc", []string{"-u", t.TunnelConfig.S6TunnelServicePath}...).Run(); err != nil {
			return err
		}

		return nil
	}

	err := os.Symlink(t.TunnelConfig.ServicePath, t.TunnelConfig.S6TunnelServicePath)
	if err != nil {
		return err
	}

	if err = exec.Command("s6-svscanctl", []string{"-a", t.TunnelConfig.S6ServicesPath}...).Run(); err != nil {
		return err
	}

	return nil
}
