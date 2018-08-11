package certificate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestFindAll(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.Mkdir(CertificatePath, os.ModeDir)

	certs := []string{"a.tld.pem", "b.tld.pem", "c.tld"}

	for _, newCert := range certs {
		fs.Create(filepath.Join("/certs", newCert))
	}

	foundCerts, err := FindAll(fs)
	if err != nil {
		t.Error(err)
	}

	if len(foundCerts) != 2 {
		t.Errorf("Unexpected cert count, got %d", len(foundCerts))
	}
}

func TestVerify(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := Verify(fs)
	if err == nil {
		t.Error("Expected error")
	}

	fs.Create("/certs/a.tld.pem")

	err = Verify(fs)
	if err != nil {
		t.Error(err)
	}
}

func TestFindForHostname(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.Create("/certs/schaper.io.pem")

	cert, err := FindForHostname("schaper.io", fs)
	if err != nil {
		t.Error(err)
	}

	if cert.Name != "schaper.io.pem" {
		t.Errorf("Unexpected cert for hostname: %s", cert.Name)
	}
}

func TestBelongsToHost(t *testing.T) {
	fs := afero.NewMemMapFs()
	cert := New("hostname.com.pem", fs)

	belongs := cert.belongsToHost("hostname.com")
	if !belongs {
		t.Errorf("Expected cert and host to belong")
	}

	belongs = cert.belongsToHost("horsename.com")
	if belongs {
		t.Errorf("Expected cert to not belong")
	}
}

func TestFullPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	name := "mysite.pem"
	cert := New(name, fs)

	expected := strings.Join([]string{CertificatePath, name}, "/")
	if cert.FullPath() != expected {
		t.Errorf("Unexpected certificate path, got %s want %s", cert.FullPath(), CertificatePath)
	}
}
