package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Tunnel struct {
	ContainerHostname string
	HeraHostname      string
	HeraPort          string
	TunnelConfig      *TunnelConfig
}

type TunnelConfig struct {
	ServicePath         string
	RunFilePath         string
	S6ServicesPath      string
	S6TunnelServicePath string
}

func NewTunnel(containerHostname string, heraHostname string, heraPort string) *Tunnel {
	servicePath := filepath.Join("/etc/services.d", heraHostname)
	runFilePath := filepath.Join(servicePath, "run")
	s6ServicesPath := "/var/run/s6/services"
	s6TunnelServicePath := filepath.Join(s6ServicesPath, heraHostname)

	tunnelConfig := &TunnelConfig{
		ServicePath:         servicePath,
		RunFilePath:         runFilePath,
		S6ServicesPath:      s6ServicesPath,
		S6TunnelServicePath: s6TunnelServicePath,
	}

	tunnel := &Tunnel{
		ContainerHostname: containerHostname,
		HeraHostname:      heraHostname,
		HeraPort:          heraPort,
		TunnelConfig:      tunnelConfig,
	}

	return tunnel
}

func (t Tunnel) Start() {
	log.Infof("Registered tunnel %s @ %s:%s", t.HeraHostname, t.ContainerHostname, t.HeraPort)
	log.Infof("Logging to /var/log/hera/%s.log", t.HeraHostname)

	t.GenerateRunFile()
	t.StartService()
}

func (t Tunnel) Stop() {
	svcArgs := []string{"-d", t.TunnelConfig.ServicePath}

	_, err := exec.Command("s6-svc", svcArgs...).Output()
	if err != nil {
		log.Errorf("Error while stopping tunnel %s: %s", t.HeraHostname, err)
		return
	}

	log.Infof("Stopped tunnel %s", t.HeraHostname)
}

func (t Tunnel) GenerateRunFile() {
	if _, err := os.Stat(t.TunnelConfig.ServicePath); os.IsNotExist(err) {
		os.Mkdir(t.TunnelConfig.ServicePath, os.ModePerm)
	}

	runFile, err := os.Create(t.TunnelConfig.RunFilePath)
	if err != nil {
		log.Error(err)
		return
	}
	defer runFile.Close()

	lines := [2]string{"#!/usr/bin/with-contenv sh\n", fmt.Sprintf("exec cloudflared --hostname %s --url %s:%s --origincert /etc/cloudflared/cert.pem --logfile /var/log/hera/%s.log", t.HeraHostname, t.ContainerHostname, t.HeraPort, t.HeraHostname)}

	writer := bufio.NewWriter(runFile)
	for _, line := range lines {
		writer.WriteString(line)
	}

	writer.Flush()

	os.Chmod(t.TunnelConfig.RunFilePath, 0755)
}

func (t Tunnel) StartService() {
	svcArgs := []string{"-u", t.TunnelConfig.S6TunnelServicePath}
	scanctlArgs := []string{"-a", t.TunnelConfig.S6ServicesPath}

	if _, err := os.Stat(t.TunnelConfig.S6TunnelServicePath); err == nil {
		_, err := exec.Command("s6-svc", svcArgs...).Output()
		if err != nil {
			log.Error(err)
		}
		return
	}

	if err := os.Symlink(t.TunnelConfig.ServicePath, t.TunnelConfig.S6TunnelServicePath); err != nil {
		log.Error(err)
		return
	}

	_, err := exec.Command("s6-svscanctl", scanctlArgs...).Output()
	if err != nil {
		log.Error(err)
	}
}
