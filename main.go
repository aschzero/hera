package main

import (
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("hera")

func main() {

	listener, err := NewListener()
	if err != nil {
		log.Errorf("Unable to start: %s", err)
	}

	log.Infof("Hera v%s has started", CurrentVersion)

	err = VerifyCertificates(listener.Fs)
	if err != nil {
		log.Error(err.Error())
	}

	err = listener.Revive()
	if err != nil {
		log.Error(err.Error())
	}

	listener.Listen()
}
