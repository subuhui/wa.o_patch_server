package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// GenerateRSAKeyPair generates a 2048-bit RSA key pair.
func GenerateRSAKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// ExportPrivateKeyPEM exports the RSA private key in PKCS#1 PEM format.
func ExportPrivateKeyPEM(privKey *rsa.PrivateKey) []byte {
	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
}

// LoadPrivateKeyPEM loads an RSA private key from PKCS#1 or PKCS#8 PEM data.
func LoadPrivateKeyPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	// Try PKCS#1 first
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try PKCS#8
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := parsedKey.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("not an RSA private key in PKCS#8 block")
	}

	return nil, fmt.Errorf("failed to parse private key: %w", err)
}

// LoadPrivateKeyFromFile loads a private key from a file path.
func LoadPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}
	return LoadPrivateKeyPEM(data)
}

// ExportPublicKeyDERBase64 exports the public key in PKIX DER format, encoded as Base64.
// This is the exact format required for `patch_public_key` in `shorebird.yaml`.
func ExportPublicKeyDERBase64(pubKey *rsa.PublicKey) (string, error) {
	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal PKIX public key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(derBytes), nil
}

// SignPatchHash signs the hex-encoded SHA256 string using RSA-PKCS1v15 with SHA-256.
// This matches the `ring::signature::RSA_PKCS1_2048_8192_SHA256` verification in Shorebird updater.
func SignPatchHash(hexHash string, privKey *rsa.PrivateKey) (string, error) {
	h := sha256.Sum256([]byte(hexHash))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, h[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign patch hash: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// VerifyPatchHash verifies the signature against the hex-encoded SHA256 string.
func VerifyPatchHash(hexHash string, sigBase64 string, pubKey *rsa.PublicKey) error {
	sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return fmt.Errorf("failed to decode signature base64: %w", err)
	}
	h := sha256.Sum256([]byte(hexHash))
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, h[:], sigBytes)
}
