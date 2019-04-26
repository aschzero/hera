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

var fs = afero.NewOsFs()

type Service struct {
	Hostname string
	Commander
}

func NewService(hostname string) *Service {
	service := &Service{
		Hostname:  hostname,
		Commander: Command{},
	}

	return service
}

func (s *Service) servicePath() string {
	return filepath.Join(ServicesPath, s.Hostname)
}

func (s *Service) ConfigFilePath() string {
	return filepath.Join(s.servicePath(), "config.yml")
}

func (s *Service) RunFilePath() string {
	return filepath.Join(s.servicePath(), "run")
}

func (s *Service) supervisePath() string {
	return filepath.Join(s.servicePath(), "supervise")
}

func (s *Service) LogFilePath() string {
	logPath := []string{filepath.Join(LogPath, s.Hostname), "log"}

	return strings.Join(logPath, ".")
}

func (s *Service) Create() error {
	exists, err := afero.DirExists(fs, s.servicePath())
	if err != nil {
		return err
	}

	if !exists {
		fs.Mkdir(s.servicePath(), os.ModePerm)
	}

	return nil
}

func (s *Service) Supervise() error {
	_, err := s.Commander.Run("s6-svscanctl", "-a", ServicesPath)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Start() error {
	_, err := s.Commander.Run("s6-svc", "-u", s.servicePath())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Stop() error {
	_, err := s.Commander.Run("s6-svc", "-d", s.servicePath())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Restart() error {
	err := s.waitUntilDown()
	if err != nil {
		return err
	}

	err = s.Start()
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

func (s *Service) IsSupervised() (bool, error) {
	registered, err := afero.DirExists(fs, s.supervisePath())
	if err != nil {
		return false, err
	}

	return registered, nil
}

func (s *Service) IsRunning() (bool, error) {
	out, err := s.Commander.Run("s6-svstat", "-u", s.servicePath())
	if err != nil {
		return false, err
	}

	return strings.Contains(string(out), "true"), nil
}
