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

// Service holds config for an s6 service
type Service struct {
	Hostname string
	Commander
}

// NewService returns a new Service. Services are used to start and stop tunnel processes,
// as well as supervise processes to ensure they are kept alive.
func NewService(hostname string) *Service {
	service := &Service{
		Hostname:  hostname,
		Commander: Command{},
	}

	return service
}

// servicePath returns the full path for the service
func (s *Service) servicePath() string {
	return filepath.Join(ServicesPath, s.Hostname)
}

// ConfigFilePath returns the full path for the service config file
func (s *Service) ConfigFilePath() string {
	return filepath.Join(s.servicePath(), "config.yml")
}

// RunFilePath returns the full path for the service run command
func (s *Service) RunFilePath() string {
	return filepath.Join(s.servicePath(), "run")
}

// supervisePath returns the full path for the service supervise command
func (s *Service) supervisePath() string {
	return filepath.Join(s.servicePath(), "supervise")
}

// LogFilePath returns the full path for the service log file
func (s *Service) LogFilePath() string {
	logPath := []string{filepath.Join(LogPath, s.Hostname), "log"}

	return strings.Join(logPath, ".")
}

// Create creates a new service directory if one does not already exist
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

// Supervise supervises a service
func (s *Service) Supervise() error {
	_, err := s.Commander.Run("s6-svscanctl", "-a", ServicesPath)
	if err != nil {
		return err
	}

	return nil
}

// Start starts a service
func (s *Service) Start() error {
	_, err := s.Commander.Run("s6-svc", "-u", s.servicePath())
	if err != nil {
		return err
	}

	return nil
}

// Stop stops a service
func (s *Service) Stop() error {
	_, err := s.Commander.Run("s6-svc", "-d", s.servicePath())
	if err != nil {
		return err
	}

	return nil
}

// Restart restarts a service
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

// waitUntilDown returns when the service is down
func (s *Service) waitUntilDown() error {
	_, err := s.Commander.Run("s6-svwait", "-d", s.servicePath())
	if err != nil {
		return err
	}

	return nil
}

// IsSupervised returns a bool to indicate if a service is supervised or not
func (s *Service) IsSupervised() (bool, error) {
	registered, err := afero.DirExists(fs, s.supervisePath())
	if err != nil {
		return false, err
	}

	return registered, nil
}

// IsRunning returns a bool to indicate if a service is running or not
func (s *Service) IsRunning() (bool, error) {
	out, err := s.Commander.Run("s6-svstat", "-u", s.servicePath())
	if err != nil {
		return false, err
	}

	return strings.Contains(string(out), "true"), nil
}
