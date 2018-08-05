package main

import (
	"os"
	"testing"

	"github.com/spf13/afero"
)

type MockCommander struct {
	mockRun func() ([]byte, error)
}

func (c MockCommander) Run(name string, arg ...string) ([]byte, error) {
	return c.mockRun()
}

func newService() *Service {
	config := &ServiceConfig{
		Hostname:       "f56540dbf360",
		TunnelHostname: "site.tld",
	}

	return NewService(config)
}

func TestServicePath(t *testing.T) {
	service := newService()

	expected := "/var/run/s6/services/site.tld"
	actual := service.servicePath()

	if actual != expected {
		t.Errorf("Unexpected service path, want %s got %s", actual, expected)
	}
}

func TestConfigFilePath(t *testing.T) {
	service := newService()

	expected := "/var/run/s6/services/site.tld/config.yml"
	actual := service.configFilePath()

	if actual != expected {
		t.Errorf("Unexpected service path, want %s got %s", actual, expected)
	}
}

func TestRunFilePath(t *testing.T) {
	service := newService()

	expected := "/var/run/s6/services/site.tld/run"
	actual := service.runFilePath()

	if actual != expected {
		t.Errorf("Unexpected run file path, want %s got %s", actual, expected)
	}
}

func TestLogFilePath(t *testing.T) {
	service := newService()

	expected := "/var/log/hera/site.tld.log"
	actual := service.logFilePath()

	if actual != expected {
		t.Errorf("Unexpected log file path, want %s got %s", actual, expected)
	}
}

func TestCreate(t *testing.T) {
	fs = afero.NewMemMapFs()
	service := newService()

	err := service.create()
	if err != nil {
		t.Error(err)
	}

	exists, err := afero.DirExists(fs, service.servicePath())
	if err != nil {
		t.Error(err)
	}

	if !exists {
		t.Error("Expected service dir")
	}
}

func TestIsSupervised(t *testing.T) {
	fs = afero.NewMemMapFs()
	service := newService()

	supervised, err := service.isSupervised()
	if err != nil {
		t.Error(err)
	}

	if supervised {
		t.Errorf("Expected service to be unsupervised")
	}

	path := service.supervisePath()
	fs.Mkdir(path, os.ModePerm)

	supervised, err = service.isSupervised()
	if err != nil {
		t.Error(err)
	}

	if !supervised {
		t.Errorf("Expected service to be supervised")
	}
}

func TestIsRunning(t *testing.T) {
	service := newService()
	service.Commander = &MockCommander{
		mockRun: func() ([]byte, error) {
			return []byte("true"), nil
		},
	}

	running, err := service.isRunning()
	if err != nil {
		t.Error(err)
	}

	if !running {
		t.Error("Service should be running")
	}

	service.Commander = &MockCommander{
		mockRun: func() ([]byte, error) {
			return []byte(""), nil
		},
	}

	running, err = service.isRunning()
	if err != nil {
		t.Error(err)
	}

	if running {
		t.Error("Service should not be running")
	}
}
