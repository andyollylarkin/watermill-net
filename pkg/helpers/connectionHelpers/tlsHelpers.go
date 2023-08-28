package connectionhelpers

import (
	"crypto/x509"

	"go.step.sm/crypto/pemutil"
)

// LoadCertPool load root certificate chain and return x509 cert pool.
func LoadCertPool(rootCAPath string) (*x509.CertPool, error) {
	cert, err := pemutil.ReadCertificate(rootCAPath)
	if err != nil {
		return nil, err
	}

	cp := x509.NewCertPool()
	cp.AddCert(cert)

	return cp, nil
}
