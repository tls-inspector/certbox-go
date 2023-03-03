package tls

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

// ExportPKCS12 will generate a PKCS12 bag for the given certificate and private key.
//
// An optional issuer certificate can be specified. When included the certificate is included in the bag.
//
// A password is required. Providing an empty string will return an error.
func ExportPKCS12(certificate *Certificate, issuer *Certificate, password string) ([]byte, error) {
	caCerts := []*x509.Certificate{}
	if issuer != nil {
		caCerts = append(caCerts, issuer.X509())
	}

	return pkcs12.Encode(rand.Reader, certificate.PKey(), certificate.X509(), caCerts, password)
}

// ExportPEM will generate PEM files for the certificate and private key.
// Returns the certificate data, key data, and optional error.
func ExportPEM(certificate *Certificate) ([]byte, []byte, error) {
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.certificateDataBytes()})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: certificate.keyDataBytes()})

	return certPEM, keyPEM, nil
}

// ExportDER will generate DER files for the certificate and private key.
// Returns the certificate data, key data, and optional error.
func ExportDER(certificate *Certificate) ([]byte, []byte, error) {
	return certificate.certificateDataBytes(), certificate.keyDataBytes(), nil
}
