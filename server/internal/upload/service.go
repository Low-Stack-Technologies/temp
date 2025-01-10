package upload

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"tech.low-stack.temp/server/internal/env"
	"tech.low-stack.temp/server/internal/storage"
	"time"
)

func Initialize() {
	http.Handle("POST /", http.HandlerFunc(handleUpload))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse multipart form:\n%s", err.Error()), http.StatusBadRequest)
		return
	}

	// Get file from request
	uploadedFile, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to get file:\n%s", err.Error()), http.StatusBadRequest)
		return
	}
	defer uploadedFile.Close()

	// Log file info
	log.Printf("Received file: %s, size: %d", header.Filename, header.Size)

	// Parse expiration duration
	expirationStr := r.FormValue("expiration")
	expiration, err := time.ParseDuration(expirationStr)
	if expirationStr == "" || expiration.Seconds() == 0 || err != nil {
		expiration = env.DefaultExpiration
	}

	// Check if expiration is valid
	if expiration < env.MinExpiration || expiration > env.MaxExpiration {
		http.Error(w, fmt.Sprintf("Invalid expiration! Must be between %s and %s", env.MinExpiration, env.MaxExpiration), http.StatusBadRequest)
		return
	}

	// Create file in database and get io writer
	file, databaseFile, err := storage.RequestNewFile(header.Filename, expiration, r.Context())
	if err != nil {
		http.Error(w, "Unable to request upload", http.StatusInternalServerError)
		return
	}

	// Copy file to storage
	_, err = io.Copy(file, uploadedFile)
	if err != nil {
		http.Error(w, "Unable to write to upload", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(databaseFile.GetDownloadURL()))
	log.Printf("Uploaded %s (%s)", databaseFile.Filename, databaseFile.ID)
}
