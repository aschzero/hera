package certificate

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/op/go-logging"

	"github.com/spf13/afero"
)

var (
	log = logging.MustGetLogger("hera")
)

const (
	CertificatePath = "/certs"
)

type Certificate struct {
	Name string
	Fs   afero.Fs
}

func New(name string, fs afero.Fs) *Certificate {
	cert := &Certificate{
		Name: name,
		Fs:   fs,
	}

	return cert
}

func FindAll(fs afero.Fs) ([]*Certificate, error) {
	var certs []*Certificate

	files, err := afero.ReadDir(fs, CertificatePath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		name := file.Name()

		if !strings.HasSuffix(name, ".pem") {
			continue
		}

		cert := New(name, fs)
		certs = append(certs, cert)
	}

	return certs, nil
}

func Verify(fs afero.Fs) error {
	certs, err := FindAll(fs)

	if err != nil || len(certs) == 0 {
		return errors.New("No certificates found")
	}

	for _, cert := range certs {
		log.Infof("Found certificate: %s", cert.Name)
	}

	return nil
}

func FindForHostname(hostname string, fs afero.Fs) (*Certificate, error) {
	certs, err := FindAll(fs)
	if err != nil {
		return nil, fmt.Errorf("Unable to scan for available certificates: %s", err)
	}

	for _, cert := range certs {
		if cert.belongsToHost(hostname) {
			return cert, nil
		}
	}

	return nil, fmt.Errorf("Unable to find certificate for %s", hostname)
}

func (c *Certificate) belongsToHost(host string) bool {
	baseCertName := strings.Split(c.Name, ".pem")[0]

	return host == baseCertName
}

func (c *Certificate) FullPath() string {
	return filepath.Join(CertificatePath, c.Name)
}

func (c *Certificate) isExist() bool {
	exists, err := afero.Exists(c.Fs, c.FullPath())
	if err != nil {
		log.Errorf("Unable to check certificate: %s", err)
	}

	return exists
}
