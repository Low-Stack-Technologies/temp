package handlers

import (
	"crypto/rsa"
	"net/http"

	"github.com/Low-Stack-Technologies/temp/server/crypto"
)

// PublicKeyHandler serves the RSA public key
type PublicKeyHandler struct {
	publicKey *rsa.PublicKey
}

// NewPublicKeyHandler creates a new public key handler
func NewPublicKeyHandler(publicKey *rsa.PublicKey) *PublicKeyHandler {
	return &PublicKeyHandler{publicKey: publicKey}
}

// ServeHTTP handles the public key request
func (h *PublicKeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Export public key as PEM
	pemBytes, err := crypto.ExportPublicKeyPEM(h.publicKey)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(pemBytes)
}
