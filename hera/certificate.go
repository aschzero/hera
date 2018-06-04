package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/radovskyb/watcher"
)

type Certificate struct {
	Directory string
	FileName  string
	Path      string
}

const (
	CertificateDirectory = "/root/.cloudflared"
	CertificateFileName  = "cert.pem"
)

const (
	CertificateIsNeededMessage = "\n Hera is unable to run without a cloudflare certificate. To fix this issue:" +
		"\n\n 1. Ensure this container has a volume mapped to `/root/.cloudflared`" +
		"\n 2. Obtain a certificate by visiting https://www.cloudflare.com/a/warp" +
		"\n 3. Rename the certificate to `cert.pem` and move it to the volume" +
		"\n\n See https://github.com/aschaper/hera#obtaining-a-certificate for more info." +
		"\n\n Hera is now watching for a `cert.pem` file and will resume operation when a certificate is found.\n"
)

func NewCertificate() *Certificate {
	certificate := &Certificate{
		Directory: CertificateDirectory,
		FileName:  CertificateFileName,
		Path:      filepath.Join(CertificateDirectory, CertificateFileName),
	}

	return certificate
}

func (c Certificate) VerifyCertificate() {
	if _, err := os.Stat(c.Path); os.IsNotExist(err) {
		log.Error(CertificateIsNeededMessage)
		c.Watch()
	}
}

func (c Certificate) Watch() {
	w := watcher.New()

	w.SetMaxEvents(1)
	w.FilterOps(watcher.Create)

	go func() {
		for {
			select {
			case event := <-w.Event:
				if event.FileInfo.Name() == c.FileName {
					log.Info("\n Found cloudflare certificate. Hera will now continue to run.\n")
					w.Close()
				}
			case err := <-w.Error:
				log.Fatal(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.Add(c.Directory); err != nil {
		log.Fatal(err)
		return
	}

	if err := w.Start(time.Millisecond * 500); err != nil {
		log.Fatal(err)
		return
	}
}
