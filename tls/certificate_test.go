package tls_test

import (
	"bytes"
	"crypto/x509"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/tls-inspector/certbox-go/tls"
)

func generateCertificateChain() (*tls.Certificate, *tls.Certificate, error) {
	root, err := tls.GenerateCertificate(tls.CertificateRequest{
		KeyType: tls.KeyTypeECDSA_256,
		Subject: tls.Name{
			Organization: "example.com",
			City:         "Vancouver",
			Province:     "British Columbia",
			Country:      "CA",
			CommonName:   "example.com Example Root",
		},
		Validity: tls.DateRange{
			NotBefore: time.Now().AddDate(-1, 0, 0),
			NotAfter:  time.Now().AddDate(20, 0, 0),
		},
		Usage: tls.KeyUsage{
			DigitalSignature: true,
			KeyEncipherment:  true,
			CRLSign:          true,
			OCSPSigning:      true,
		},
		IsCertificateAuthority: true,
		SignatureAlgorithm:     tls.SignatureAlgorithmSHA256,
	}, nil)
	if err != nil {
		return nil, nil, err
	}

	leaf, err := tls.GenerateCertificate(tls.CertificateRequest{
		KeyType: tls.KeyTypeECDSA_256,
		Subject: tls.Name{
			Organization: "example.com",
			City:         "Vancouver",
			Province:     "British Columbia",
			Country:      "CA",
			CommonName:   "foo.example.com",
		},
		AlternateNames: []tls.AlternateName{
			{
				Type:  tls.AlternateNameTypeDNS,
				Value: "foo.example.com",
			},
			{
				Type:  tls.AlternateNameTypeDNS,
				Value: "*.foo.example.com",
			},
		},
		Validity: tls.DateRange{
			NotBefore: time.Now().AddDate(0, 0, -1),
			NotAfter:  time.Now().AddDate(0, 0, 350),
		},
		Usage: tls.KeyUsage{
			DigitalSignature: true,
			ServerAuth:       true,
			ClientAuth:       true,
		},
		SignatureAlgorithm: tls.SignatureAlgorithmSHA256,
	}, root)
	if err != nil {
		return nil, nil, err
	}

	return root, leaf, err
}

func TestGenerateCertificate(t *testing.T) {
	t.Parallel()

	root, leaf, err := generateCertificateChain()

	if err != nil {
		t.Fatalf("Error generating certificate chain: %s", err.Error())
	}

	rootCert := root.X509()
	leafCert := leaf.X509()

	if strings.Join(leafCert.Issuer.Country, "") != strings.Join(rootCert.Subject.Country, "") {
		t.Errorf("Leaf issuer does not match root: '%s' != '%s'", leafCert.Issuer.Country, leafCert.Subject.Country)
	}
	if strings.Join(leafCert.Issuer.Organization, "") != strings.Join(rootCert.Subject.Organization, "") {
		t.Errorf("Leaf issuer does not match root: '%s' != '%s'", leafCert.Issuer.Organization, leafCert.Subject.Organization)
	}
	if strings.Join(leafCert.Issuer.OrganizationalUnit, "") != strings.Join(rootCert.Subject.OrganizationalUnit, "") {
		t.Errorf("Leaf issuer does not match root: '%s' != '%s'", leafCert.Issuer.OrganizationalUnit, leafCert.Subject.OrganizationalUnit)
	}
	if strings.Join(leafCert.Issuer.Locality, "") != strings.Join(rootCert.Subject.Locality, "") {
		t.Errorf("Leaf issuer does not match root: '%s' != '%s'", leafCert.Issuer.Locality, leafCert.Subject.Locality)
	}
	if strings.Join(leafCert.Issuer.Province, "") != strings.Join(rootCert.Subject.Province, "") {
		t.Errorf("Leaf issuer does not match root: '%s' != '%s'", leafCert.Issuer.Province, leafCert.Subject.Province)
	}
	if strings.Join(leafCert.Issuer.StreetAddress, "") != strings.Join(rootCert.Subject.StreetAddress, "") {
		t.Errorf("Leaf issuer does not match root: '%s' != '%s'", leafCert.Issuer.StreetAddress, leafCert.Subject.StreetAddress)
	}
	if strings.Join(leafCert.Issuer.PostalCode, "") != strings.Join(rootCert.Subject.PostalCode, "") {
		t.Errorf("Leaf issuer does not match root: '%s' != '%s'", leafCert.Issuer.PostalCode, leafCert.Subject.PostalCode)
	}
	if leafCert.Issuer.SerialNumber != rootCert.Subject.SerialNumber {
		t.Errorf("Leaf issuer does not match root: '%s' != '%s'", leafCert.Issuer.SerialNumber, leafCert.Subject.SerialNumber)
	}
	if leafCert.Issuer.CommonName != rootCert.Subject.CommonName {
		t.Errorf("Leaf issuer does not match root: '%s' != '%s'", leafCert.Issuer.CommonName, leafCert.Subject.CommonName)
	}

	if !bytes.Equal(leafCert.AuthorityKeyId, rootCert.SubjectKeyId) {
		t.Errorf("AKID does not match root SKID")
	}

	if rootCert.AuthorityKeyId != nil {
		t.Errorf("Root cert should not have AKID")
	}
}

func TestDateRange(t *testing.T) {
	t.Parallel()

	dateRange := tls.DateRange{
		NotBefore: time.Now().AddDate(0, -1, 0),
		NotAfter:  time.Now().AddDate(0, 1, 0),
	}
	if !dateRange.IsValid() {
		t.Fatalf("Incorrect date range valid result for valid range")
	}

	dateRange = tls.DateRange{
		NotBefore: time.Now().AddDate(0, 1, 0),
		NotAfter:  time.Now().AddDate(0, 2, 0),
	}
	if dateRange.IsValid() {
		t.Fatalf("Incorrect date range valid result future range")
	}

	dateRange = tls.DateRange{
		NotBefore: time.Now().AddDate(0, -2, 0),
		NotAfter:  time.Now().AddDate(0, -1, 0),
	}
	if dateRange.IsValid() {
		t.Fatalf("Incorrect date range valid result past range")
	}
}

func TestSAN(t *testing.T) {
	t.Parallel()

	dnsName := "*.example.com"
	email := "hello@example.com"
	ipAddress := "127.0.0.1"
	uri := "http://user:pass@example.com:80/index.html#anchor?foo=bar"

	cert, err := tls.GenerateCertificate(tls.CertificateRequest{
		KeyType: tls.KeyTypeRSA_4096,
		Subject: tls.Name{
			Organization: "example.com",
			City:         "Vancouver",
			Province:     "British Columbia",
			Country:      "CA",
			CommonName:   "example.com Example Root",
		},
		Validity: tls.DateRange{
			NotBefore: time.Now().AddDate(-1, 0, 0),
			NotAfter:  time.Now().AddDate(20, 0, 0),
		},
		AlternateNames: []tls.AlternateName{
			{
				Type:  tls.AlternateNameTypeDNS,
				Value: dnsName,
			},
			{
				Type:  tls.AlternateNameTypeEmail,
				Value: email,
			},
			{
				Type:  tls.AlternateNameTypeIP,
				Value: ipAddress,
			},
			{
				Type:  tls.AlternateNameTypeURI,
				Value: uri,
			},
		},
		IsCertificateAuthority: true,
		SignatureAlgorithm:     tls.SignatureAlgorithmSHA256,
	}, nil)
	if err != nil {
		t.Fatalf("Error generating certificate: %s", err.Error())
	}

	data, err := hex.DecodeString(cert.CertificateData)
	if err != nil {
		t.Fatalf("Error decoding certificate bytes: %s", err.Error())
	}

	x, err := x509.ParseCertificate(data)
	if err != nil {
		t.Fatalf("Error parsing X.509 data: %s", err.Error())
	}

	if len(x.DNSNames) == 0 {
		t.Errorf("Certificate has no DNSNames")
	}
	if len(x.URIs) == 0 {
		t.Errorf("Certificate has no URIs")
	}
	if len(x.EmailAddresses) == 0 {
		t.Errorf("Certificate has no EmailAddresses")
	}
	if len(x.IPAddresses) == 0 {
		t.Errorf("Certificate has no IPAddresses")
	}

	if x.DNSNames[0] != dnsName {
		t.Errorf("Unexpected value for DNS altername name. Expected '%s' got '%s'", dnsName, x.DNSNames[0])
	}
	if x.URIs[0].String() != uri {
		t.Errorf("Unexpected value for URI altername name. Expected '%s' got '%s'", uri, x.URIs[0].String())
	}
	if x.EmailAddresses[0] != email {
		t.Errorf("Unexpected value for EMail altername name. Expected '%s' got '%s'", email, x.EmailAddresses[0])
	}
	if x.IPAddresses[0].String() != ipAddress {
		t.Errorf("Unexpected value for IPAddress altername name. Expected '%s' got '%s'", ipAddress, x.IPAddresses[0].String())
	}
}

func TestSANInvalidTypes(t *testing.T) {
	// Empty value
	_, err := tls.GenerateCertificate(tls.CertificateRequest{
		KeyType: tls.KeyTypeECDSA_256,
		Subject: tls.Name{
			Organization: "example.com",
			City:         "Vancouver",
			Province:     "British Columbia",
			Country:      "CA",
			CommonName:   "example.com Example Root",
		},
		Validity: tls.DateRange{
			NotBefore: time.Now().AddDate(-1, 0, 0),
			NotAfter:  time.Now().AddDate(20, 0, 0),
		},
		AlternateNames: []tls.AlternateName{
			{
				Type: tls.AlternateNameTypeDNS,
			},
		},
		IsCertificateAuthority: true,
		SignatureAlgorithm:     tls.SignatureAlgorithmSHA256,
	}, nil)
	if err == nil {
		t.Errorf("No error seen when one expected for empty value")
	}

	// Invalid DNS name
	_, err = tls.GenerateCertificate(tls.CertificateRequest{
		KeyType: tls.KeyTypeECDSA_256,
		Subject: tls.Name{
			Organization: "example.com",
			City:         "Vancouver",
			Province:     "British Columbia",
			Country:      "CA",
			CommonName:   "example.com Example Root",
		},
		Validity: tls.DateRange{
			NotBefore: time.Now().AddDate(-1, 0, 0),
			NotAfter:  time.Now().AddDate(20, 0, 0),
		},
		AlternateNames: []tls.AlternateName{
			{
				Type:  tls.AlternateNameTypeDNS,
				Value: " ",
			},
		},
		IsCertificateAuthority: true,
		SignatureAlgorithm:     tls.SignatureAlgorithmSHA256,
	}, nil)
	if err == nil {
		t.Errorf("No error seen when one expected for invalid DNS name")
	}

	// Invalid IP
	_, err = tls.GenerateCertificate(tls.CertificateRequest{
		KeyType: tls.KeyTypeECDSA_256,
		Subject: tls.Name{
			Organization: "example.com",
			City:         "Vancouver",
			Province:     "British Columbia",
			Country:      "CA",
			CommonName:   "example.com Example Root",
		},
		Validity: tls.DateRange{
			NotBefore: time.Now().AddDate(-1, 0, 0),
			NotAfter:  time.Now().AddDate(20, 0, 0),
		},
		AlternateNames: []tls.AlternateName{
			{
				Type:  tls.AlternateNameTypeIP,
				Value: "256.256.256.256.256",
			},
		},
		IsCertificateAuthority: true,
		SignatureAlgorithm:     tls.SignatureAlgorithmSHA256,
	}, nil)
	if err == nil {
		t.Errorf("No error seen when one expected for invalid IP address")
	}

	// Invalid URI
	_, err = tls.GenerateCertificate(tls.CertificateRequest{
		KeyType: tls.KeyTypeECDSA_256,
		Subject: tls.Name{
			Organization: "example.com",
			City:         "Vancouver",
			Province:     "British Columbia",
			Country:      "CA",
			CommonName:   "example.com Example Root",
		},
		Validity: tls.DateRange{
			NotBefore: time.Now().AddDate(-1, 0, 0),
			NotAfter:  time.Now().AddDate(20, 0, 0),
		},
		AlternateNames: []tls.AlternateName{
			{
				Type:  tls.AlternateNameTypeURI,
				Value: "http://[::1]:namedport",
			},
		},
		IsCertificateAuthority: true,
		SignatureAlgorithm:     tls.SignatureAlgorithmSHA256,
	}, nil)
	if err == nil {
		t.Errorf("No error seen when one expected for invalid URI")
	}

	// Invalid san type
	_, err = tls.GenerateCertificate(tls.CertificateRequest{
		KeyType: tls.KeyTypeECDSA_256,
		Subject: tls.Name{
			Organization: "example.com",
			City:         "Vancouver",
			Province:     "British Columbia",
			Country:      "CA",
			CommonName:   "example.com Example Root",
		},
		Validity: tls.DateRange{
			NotBefore: time.Now().AddDate(-1, 0, 0),
			NotAfter:  time.Now().AddDate(20, 0, 0),
		},
		AlternateNames: []tls.AlternateName{
			{
				Type:  "foo",
				Value: "bar",
			},
		},
		IsCertificateAuthority: true,
		SignatureAlgorithm:     tls.SignatureAlgorithmSHA256,
	}, nil)
	if err == nil {
		t.Errorf("No error seen when one expected for unknown name type")
	}
}

func TestKeyUsage(t *testing.T) {
	t.Parallel()

	cert, err := tls.GenerateCertificate(tls.CertificateRequest{
		KeyType: tls.KeyTypeECDSA_256,
		Subject: tls.Name{
			Organization: "example.com",
			City:         "Vancouver",
			Province:     "British Columbia",
			Country:      "CA",
			CommonName:   "example.com Example Root",
		},
		Validity: tls.DateRange{
			NotBefore: time.Now().AddDate(-1, 0, 0),
			NotAfter:  time.Now().AddDate(20, 0, 0),
		},
		Usage: tls.KeyUsage{
			DigitalSignature:  true,
			ContentCommitment: true,
			KeyEncipherment:   true,
			DataEncipherment:  true,
			KeyAgreement:      true,
			CertSign:          true,
			CRLSign:           true,
			EncipherOnly:      true,
			DecipherOnly:      true,
			ServerAuth:        true,
			ClientAuth:        true,
			CodeSigning:       true,
			EmailProtection:   true,
			TimeStamping:      true,
			OCSPSigning:       true,
		},
		IsCertificateAuthority: true,
		SignatureAlgorithm:     tls.SignatureAlgorithmSHA256,
	}, nil)
	if err != nil {
		t.Fatalf("Error generating certificate: %s", err.Error())
	}

	data, err := hex.DecodeString(cert.CertificateData)
	if err != nil {
		t.Fatalf("Error decoding certificate bytes: %s", err.Error())
	}

	x, err := x509.ParseCertificate(data)
	if err != nil {
		t.Fatalf("Error parsing X.509 data: %s", err.Error())
	}

	if x.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		t.Errorf("Certificate does not have DigitalSignature usage")
	}
	if x.KeyUsage&x509.KeyUsageContentCommitment == 0 {
		t.Errorf("Certificate does not have ContentCommitment usage")
	}
	if x.KeyUsage&x509.KeyUsageKeyEncipherment == 0 {
		t.Errorf("Certificate does not have KeyEncipherment usage")
	}
	if x.KeyUsage&x509.KeyUsageDataEncipherment == 0 {
		t.Errorf("Certificate does not have DataEncipherment usage")
	}
	if x.KeyUsage&x509.KeyUsageKeyAgreement == 0 {
		t.Errorf("Certificate does not have KeyAgreement usage")
	}
	if x.KeyUsage&x509.KeyUsageCertSign == 0 {
		t.Errorf("Certificate does not have CertSign usage")
	}
	if x.KeyUsage&x509.KeyUsageCRLSign == 0 {
		t.Errorf("Certificate does not have CRLSign usage")
	}
	if x.KeyUsage&x509.KeyUsageEncipherOnly == 0 {
		t.Errorf("Certificate does not have EncipherOnly usage")
	}
	if x.KeyUsage&x509.KeyUsageDecipherOnly == 0 {
		t.Errorf("Certificate does not have DecipherOnly usage")
	}

	hasExtKeyUsageServerAuth := false
	hasExtKeyUsageClientAuth := false
	hasExtKeyUsageCodeSigning := false
	hasExtKeyUsageEmailProtection := false
	hasExtKeyUsageTimeStamping := false
	hasExtKeyUsageOCSPSigning := false

	for _, usage := range x.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth {
			hasExtKeyUsageServerAuth = true
		} else if usage == x509.ExtKeyUsageClientAuth {
			hasExtKeyUsageClientAuth = true
		} else if usage == x509.ExtKeyUsageCodeSigning {
			hasExtKeyUsageCodeSigning = true
		} else if usage == x509.ExtKeyUsageEmailProtection {
			hasExtKeyUsageEmailProtection = true
		} else if usage == x509.ExtKeyUsageTimeStamping {
			hasExtKeyUsageTimeStamping = true
		} else if usage == x509.ExtKeyUsageOCSPSigning {
			hasExtKeyUsageOCSPSigning = true
		}
	}

	if !hasExtKeyUsageServerAuth {
		t.Errorf("Certificate does not have extended ServerAuth usage")
	}
	if !hasExtKeyUsageClientAuth {
		t.Errorf("Certificate does not have extended ClientAuth usage")
	}
	if !hasExtKeyUsageCodeSigning {
		t.Errorf("Certificate does not have extended CodeSigning usage")
	}
	if !hasExtKeyUsageEmailProtection {
		t.Errorf("Certificate does not have extended EmailProtection usage")
	}
	if !hasExtKeyUsageTimeStamping {
		t.Errorf("Certificate does not have extended TimeStamping usage")
	}
	if !hasExtKeyUsageOCSPSigning {
		t.Errorf("Certificate does not have extended OCSPSigning usage")
	}
}
