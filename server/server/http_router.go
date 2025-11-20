package server

import (
	"crypto/rsa"
	"net/http"

	"github.com/Low-Stack-Technologies/temp/server/config"
	"github.com/Low-Stack-Technologies/temp/server/db"
	"github.com/Low-Stack-Technologies/temp/server/handlers"
	"github.com/Low-Stack-Technologies/temp/server/storage"
)

// SetupHTTPRouter creates and configures the HTTP router
func SetupHTTPRouter(
	cfg *config.Config,
	repo *db.Repository,
	stor *storage.Storage,
	publicKey *rsa.PublicKey,
) *http.ServeMux {
	mux := http.NewServeMux()
	
	// Upload endpoint
	uploadHandler := handlers.NewHTTPUploadHandler(cfg, repo, stor)
	mux.Handle("/api/upload", uploadHandler)
	
	// Download endpoint
	downloadHandler := handlers.NewFileDownloadHandler(repo, stor)
	mux.HandleFunc("/api/download/", downloadHandler.ServeHTTP)
	
	// Public key endpoint
	publicKeyHandler := handlers.NewPublicKeyHandler(publicKey)
	mux.Handle("/api/public-key", publicKeyHandler)
	
	// File listing endpoint
	listHandler := handlers.NewFileListHandler(cfg, repo)
	mux.Handle("/api/files", listHandler)
	
	// File metadata endpoint
	metadataHandler := handlers.NewFileMetadataHandler(cfg, repo)
	mux.HandleFunc("/api/files/", func(w http.ResponseWriter, r *http.Request) {
		// Only handle if it's not the list endpoint
		if r.URL.Path == "/api/files" || r.URL.Path == "/api/files/" {
			listHandler.ServeHTTP(w, r)
		} else {
			metadataHandler.ServeHTTP(w, r)
		}
	})
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	return mux
}
