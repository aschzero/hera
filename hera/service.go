package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

const (
	ServicesPath = "/var/run/s6/services"
	LogPath      = "/var/log/hera"
)

type Service struct {
	Config *ServiceConfig
	Commander
}

type ServiceConfig struct {
	Hostname       string
	TunnelHostname string
}

func NewService(config *ServiceConfig) *Service {
	service := &Service{
		Config:    config,
		Commander: Command{},
	}

	return service
}

func (s *Service) servicePath() string {
	return filepath.Join(ServicesPath, s.Config.TunnelHostname)
}

func (s *Service) configFilePath() string {
	return filepath.Join(s.servicePath(), "config.yml")
}

func (s *Service) runFilePath() string {
	return filepath.Join(s.servicePath(), "run")
}

func (s *Service) supervisePath() string {
	return filepath.Join(s.servicePath(), "supervise")
}

func (s *Service) logFilePath() string {
	logPath := []string{filepath.Join(LogPath, s.Config.TunnelHostname), "log"}

	return strings.Join(logPath, ".")
}

func (s *Service) create() error {
	exists, err := afero.DirExists(fs, s.servicePath())
	if err != nil {
		return err
	}

	if !exists {
		fs.Mkdir(s.servicePath(), os.ModePerm)
	}

	return nil
}

func (s *Service) supervise() error {
	_, err := s.Commander.Run("s6-svscanctl", "-a", ServicesPath)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) start() error {
	_, err := s.Commander.Run("s6-svc", "-u", s.servicePath())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) stop() error {
	_, err := s.Commander.Run("s6-svc", "-d", s.servicePath())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) restart() error {
	err := s.waitUntilDown()
	if err != nil {
		return err
	}

	_, err = s.Commander.Run("s6-svc", "-wr", s.servicePath())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) waitUntilDown() error {
	_, err := s.Commander.Run("s6-svwait", "-d", s.servicePath())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) isSupervised() (bool, error) {
	registered, err := afero.DirExists(fs, s.supervisePath())
	if err != nil {
		return false, err
	}

	return registered, nil
}

func (s *Service) isRunning() (bool, error) {
	out, err := s.Commander.Run("s6-svstat", "-u", s.servicePath())
	if err != nil {
		return false, err
	}

	return strings.Contains(string(out), "true"), nil
}
