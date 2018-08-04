package main

import "github.com/spf13/afero"

var fs = afero.NewOsFs()

func main() {
	initLogger()
	log.Infof("Hera v%s has started", Version)

	certConfig := NewCertificateConfig()
	err := certConfig.checkCertificates()
	if err != nil {
		log.Error(err)
	}

	run()
}
