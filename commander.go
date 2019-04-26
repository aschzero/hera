package main

import (
	"os/exec"
)

type Commander interface {
	Run(name string, arg ...string) ([]byte, error)
}
type Command struct{}

func (c Command) Run(name string, arg ...string) ([]byte, error) {
	out, err := exec.Command(name, arg...).Output()
	return out, err
}
