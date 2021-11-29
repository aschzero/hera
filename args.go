package main

import (
	"github.com/alexflint/go-arg"
)

var conf struct {
	Swarm bool `arg:"env:HERA_SWARM" help:"enable Swarm mode"`
	// Network       string `arg:"--net,-n,env:HERA_NETWORK" help:"docker network name to monitor" default:"hera"`
	HostnameLabel string `arg:"--host,-l,env:HERA_HOST_LABEL" help:"label containing public tunnel host" default:"hera.hostname"`
	PortLabel     string `arg:"--port,-p,env:HERA_PORT_LABEL" help:"label containing the container/service port" default:"hera.port"`
}

var (
	swarmMode     bool
	heraHostLabel string
	heraPortLabel string
	heraNetwork   string
)

func init() {
	InitLogger("hera.log")
	arg.MustParse(&conf)

	swarmMode = conf.Swarm
	heraNetwork = conf.PortLabel
	heraHostLabel = conf.HostnameLabel
	heraPortLabel = conf.PortLabel

	mode := "standalone"
	if swarmMode {
		mode = "swarm"
	}

	log.Infof("starting hera in %s mode", mode)
}
