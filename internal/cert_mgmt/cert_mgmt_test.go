package cert_mgmt

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	utls "github.com/refraction-networking/utls"
	"gotest.tools/assert"
)

func TestLoadX509KeyPair_NewCertificate(t *testing.T) {
	// Setup: define file paths for test
	certFile := "test_cert.pem"
	keyFile := "test_key.pem"

	// Ensure any existing test files are removed
	os.Remove(certFile)
	os.Remove(keyFile)

	// Test: Load or create the key pair
	_, err := LoadX509KeyPair(certFile, keyFile)
	assert.NilError(t, err, "should not error when creating a new certificate")

	// Check if files were created
	_, err = os.Stat(certFile)
	assert.NilError(t, err, "certificate file should exist")

	_, err = os.Stat(keyFile)
	assert.NilError(t, err, "key file should exist")

	// Cleanup
	os.Remove(certFile)
	os.Remove(keyFile)
}

func TestLoadX509KeyPair_ExpiredCertificate(t *testing.T) {
	// Setup: create an expired certificate
	certFile := "expired_test_cert.pem"
	keyFile := "expired_test_key.pem"
	createExpiredCertificate(certFile, keyFile)

	// Test: Load or create the key pair
	_, err := LoadX509KeyPair(certFile, keyFile)
	assert.NilError(t, err, "should not error when replacing an expired certificate")

	// Verify that a new certificate was generated file exists
	_, err = os.Stat(certFile)
	assert.NilError(t, err, "certificate file should exist")

	// Read the new certificate file
	certFileDecoded, err := os.ReadFile(certFile)
	assert.NilError(t, err, "should be able to read the new certificate file")

	// Decode the certificate
	block, _ := pem.Decode(certFileDecoded)
	assert.Assert(t, block != nil, "should be able to decode the new certificate file")

	// Parse the certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	assert.NilError(t, err, "should be able to parse the new certificate file")

	// Verify that the new certificate is not expired
	assert.Assert(t, time.Now().Before(cert.NotAfter), "certificate should not be expired")
	// Cleanup
	os.Remove(certFile)
	os.Remove(keyFile)
}

// createExpiredCertificate is a helper function to generate an expired certificate for testing purposes
func createExpiredCertificate(certFile, keyFile string) (utls.Certificate, error) {
	// Call 	 with negative validity to create an expired certificate
	return GenerateKeyPair(certFile, keyFile, -48*time.Hour) // Negative duration for expiration

}
