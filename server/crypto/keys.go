package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// KeyPair holds RSA public and private keys
type KeyPair struct {
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
}

// GenerateKeyPair generates a new RSA-2048 key pair
func GenerateKeyPair() (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	
	return &KeyPair{
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	}, nil
}

// SavePrivateKey saves the private key to a file in PEM format
func SavePrivateKey(keyPath string, privateKey *rsa.PrivateKey) error {
	// Ensure directory exists
	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}
	
	// Encode private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	
	// Write to file with restricted permissions
	file, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer file.Close()
	
	if err := pem.Encode(file, privateKeyPEM); err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}
	
	return nil
}

// LoadPrivateKey loads a private key from a file
func LoadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}
	
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	
	return privateKey, nil
}

// LoadOrGenerateKeyPair loads an existing key pair or generates a new one
func LoadOrGenerateKeyPair(keyPath string) (*KeyPair, error) {
	// Try to load existing key
	if _, err := os.Stat(keyPath); err == nil {
		privateKey, err := LoadPrivateKey(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load existing key: %w", err)
		}
		return &KeyPair{
			Private: privateKey,
			Public:  &privateKey.PublicKey,
		}, nil
	}
	
	// Generate new key pair
	keyPair, err := GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	
	// Save private key
	if err := SavePrivateKey(keyPath, keyPair.Private); err != nil {
		return nil, err
	}
	
	return keyPair, nil
}

// ExportPublicKeyPEM exports the public key in PEM format
func ExportPublicKeyPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	
	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	
	return pem.EncodeToMemory(publicKeyPEM), nil
}
