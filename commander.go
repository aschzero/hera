package main

import (
	"os/exec"
)

// Commander represents an interface for exec commands
type Commander interface {
	Run(name string, arg ...string) ([]byte, error)
}

type Command struct{}

// Run executes a command returns the output
func (c Command) Run(name string, arg ...string) ([]byte, error) {
	out, err := exec.Command(name, arg...).Output()
	return out, err
}
