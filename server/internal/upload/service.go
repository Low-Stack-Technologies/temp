package upload

import (
	"fmt"
	"github.com/dustin/go-humanize"
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
		log.Printf("Unable to parse multipart form:\n%s", err.Error())
		http.Error(w, "Unable to parse multipart form!", http.StatusInternalServerError)
		return
	}

	// Get file from request
	uploadedFile, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Unable to form file:\n%s", err.Error())
		http.Error(w, fmt.Sprintf("Unable to get file:\n%s", err.Error()), http.StatusBadRequest)
		return
	}
	defer uploadedFile.Close()

	// Log file info
	log.Printf("Received file: %s, size: %s", header.Filename, humanize.Bytes(uint64(header.Size)))

	// Ensure file is not over maximum allowed size
	if uint64(header.Size) > env.MaxFileSize {
		log.Printf("File is too large! %s > %s", humanize.Bytes(uint64(header.Size)), humanize.Bytes(env.MaxFileSize))
		http.Error(w, fmt.Sprintf("File is too large! Uploaded file is %s, max allowed is %s.", humanize.Bytes(uint64(header.Size)), humanize.Bytes(env.MaxFileSize)), http.StatusBadRequest)
		return
	}

	// Ensure free space is sufficient
	freeSpace, err := storage.GetFreeSpace()
	if err != nil {
		log.Printf("Unable to get free space: %s", err.Error())
		http.Error(w, fmt.Sprintf("Unable to get free space:\n%s", err.Error()), http.StatusInternalServerError)
		return
	}
	if freeSpace < env.MinFreeSpace+uint64(header.Size) {
		log.Printf("Insufficient free space! Needed: %s, available: %s", humanize.Bytes(env.MinFreeSpace+uint64(header.Size)), humanize.Bytes(freeSpace))
		http.Error(w, "Insufficient free space!", http.StatusInsufficientStorage)
		return
	}

	// Parse expiration duration
	expirationStr := r.FormValue("expiration")
	expiration, err := time.ParseDuration(expirationStr)
	if expirationStr == "" || expiration.Seconds() == 0 || err != nil {
		expiration = env.DefaultExpiration
	}

	// Check if expiration is valid
	if expiration < env.MinExpiration || expiration > env.MaxExpiration {
		log.Printf("Invalid expiration! Must be between %s and %s", env.MinExpiration, env.MaxExpiration)
		http.Error(w, fmt.Sprintf("Invalid expiration! Must be between %s and %s", env.MinExpiration, env.MaxExpiration), http.StatusBadRequest)
		return
	}

	// Create file in database and get io writer
	file, databaseFile, err := storage.RequestNewFile(header.Filename, expiration, r.Context())
	if err != nil {
		log.Printf("Unable to request upload: %s", err.Error())
		http.Error(w, "Unable to request upload", http.StatusInternalServerError)
		return
	}

	// Copy file to storage
	_, err = io.Copy(file, uploadedFile)
	if err != nil {
		log.Printf("Unable to write to upload: %s", err.Error())
		http.Error(w, "Unable to write to upload", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(databaseFile.GetDownloadURL()))
	log.Printf("Uploaded %s (%s)", databaseFile.Filename, databaseFile.ID)
}
