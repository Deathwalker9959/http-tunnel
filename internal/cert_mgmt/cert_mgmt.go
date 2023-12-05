package cert_mgmt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	utls "github.com/refraction-networking/utls"
)

const CERT_VALIDITY_PERIOD = 100 * 365 * 24 * time.Hour

// LoadX509KeyPair checks for existing X509 key pair, checks if it's expired, and creates a new one if necessary.
func LoadX509KeyPair(certFile, keyFile string) (utls.Certificate, error) {
	// Check if the certificate file exists
	if _, err := os.Stat(certFile); err == nil {
		// Read and parse the certificate
		certPEM, err := os.ReadFile(certFile)
		if err != nil {
			return utls.Certificate{}, err
		}

		block, _ := pem.Decode(certPEM)
		if block == nil {
			return utls.Certificate{}, err
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return utls.Certificate{}, err
		}

		// Check if the certificate is expired
		if time.Now().After(cert.NotAfter) {
			// Certificate is expired, regenerate key pair
			return GenerateKeyPair(certFile, keyFile, CERT_VALIDITY_PERIOD)
		}

		// Certificate is not expired, load and return it
		return utls.LoadX509KeyPair(certFile, keyFile)
	}

	// Certificate file does not exist or is unreadable, generate new key pair
	return GenerateKeyPair(certFile, keyFile, CERT_VALIDITY_PERIOD)
}

// generateKeyPair generates a new X509 key pair.
func GenerateKeyPair(certFile, keyFile string, validity time.Duration) (utls.Certificate, error) {
	// Generate a new private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return utls.Certificate{}, err
	}

	// Create a certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(CERT_VALIDITY_PERIOD)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return utls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Internet Widgits Pty Ltd"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add localhost as IP SAN
	ipAddresses := []net.IP{net.ParseIP("127.0.0.1")}
	template.IPAddresses = ipAddresses

	// Create a self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return utls.Certificate{}, err
	}

	// Check if the path exists and contains more than one part
	if _, err := os.Stat(certFile); os.IsNotExist(err) && strings.Count(certFile, "/") > 1 {
		// Extract the parent directory of the path
		parentDir := certFile[:strings.LastIndex(certFile, "/")]

		// Create the parent directory
		if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
			// Handle the error if mkdir fails
			panic(err)
		}
	}

	// Save the certificate
	certOut, err := os.OpenFile(certFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return utls.Certificate{}, err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	// Save the private key
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return utls.Certificate{}, err
	}

	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return utls.Certificate{}, err
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	keyOut.Close()

	// Load the newly created key pair into a utls.Certificate
	cert, err := utls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return utls.Certificate{}, err
	}

	return cert, nil
}
