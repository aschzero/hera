package main

import (
	"github.com/aschzero/hera/certificate"
	"github.com/aschzero/hera/listener"
	"github.com/aschzero/hera/logger"
	"github.com/aschzero/hera/version"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("hera")

func main() {
	logger.Init("hera")

	listener, err := listener.New()
	if err != nil {
		log.Errorf("Unable to start: %s", err)
	}

	log.Infof("Hera v%s has started", version.Current)

	err = certificate.Verify(listener.Fs)
	if err != nil {
		log.Error(err.Error())
	}

	err = listener.Revive()
	if err != nil {
		log.Error(err.Error())
	}

	listener.Listen()
}
