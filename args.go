package main

import (
	"github.com/alexflint/go-arg"
)

var heraArgs struct {
	Swarm         bool   `arg:"env:HERA_SWARM" help:"enable Swarm mode"`
	Network       string `arg:"--net,-n,env:HERA_NETWORK" help:"docker network name to monitor" default:"hera"`
	HostnameLabel string `arg:"--host,-l,env:HERA_HOSTNAME" help:"label containing public tunnel host" default:"hera.hostname"`
	PortLabel     string `arg:"--port,-p,env:HERA_PORT" help:"label containing the container/service port" default:"hera.port"`
}

var (
	SwarmMode    bool
	heraNetwork  string
	heraHostname string
	heraPort     string
)

func init() {
	InitLogger("hera.log")
	arg.MustParse(&heraArgs)

	SwarmMode = heraArgs.Swarm
	heraNetwork = heraArgs.PortLabel
	heraHostname = heraArgs.HostnameLabel
	heraPort = heraArgs.PortLabel

	mode := "engine"
	if SwarmMode {
		mode = "swarm"
	}

	log.Infof("starting hera in %s mode", mode)
}
