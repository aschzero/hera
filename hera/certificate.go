package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/spf13/afero"
)

// Certificate holds metadata of the cert.pem file
type Certificate struct {
	Directory string
	FileName  string
	Path      string
}

// NewCertificate returns a Certificate with default metadata
func NewCertificate() *Certificate {
	dir := "/root/.cloudflared"
	name := "cert.pem"

	certificate := &Certificate{
		Directory: dir,
		FileName:  name,
		Path:      filepath.Join(dir, name),
	}

	return certificate
}

// VerifyCertificate verifies the presence of a certificate
func (c Certificate) VerifyCertificate() error {
	exists, err := afero.Exists(fs, c.Path)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%s not found", c.Path)
	}

	return nil
}

// Wait pauses Hera until a certificate file is found
func (c Certificate) Wait() {
	w := watcher.New()

	w.SetMaxEvents(1)
	w.FilterOps(watcher.Create)

	go func() {
		for {
			select {
			case event := <-w.Event:
				if event.FileInfo.Name() == c.FileName {
					log.Info("Found cloudflare certificate. Hera will now resume.\n")
					w.Close()
				}
			case err := <-w.Error:
				log.Info(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.Add(c.Directory); err != nil {
		log.Error(err)
	}

	if err := w.Start(time.Millisecond * 500); err != nil {
		log.Fatal(err)
		return
	}
}

// CertificateIsNeededMessage is displayed when a cert.pem file cannot be found
const (
	CertificateIsNeededMessage = "\n Hera is unable to run without a cloudflare certificate. To fix this issue:" +
		"\n\n 1. Ensure this container has a volume mapped to `/root/.cloudflared`" +
		"\n 2. Obtain a certificate by visiting https://www.cloudflare.com/a/warp" +
		"\n 3. Rename the certificate to `cert.pem` and move it to the volume" +
		"\n\n See https://github.com/aschzero/hera#obtain-a-certificate for more info." +
		"\n\n Hera is now watching for a `cert.pem` file and will resume operation when a certificate is found.\n"
)
