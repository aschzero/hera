package main

import (
	"testing"

	"github.com/spf13/afero"
)

func TestVerifyCertificate(t *testing.T) {
	fs = afero.NewMemMapFs()
	cert := NewCertificate()

	if err := cert.VerifyCertificate(); err == nil {
		t.Error("want error")
	}

	if _, err := fs.Create(cert.Path); err != nil {
		t.Error(err)
	}

	if err := cert.VerifyCertificate(); err != nil {
		t.Error(err)
	}
}
