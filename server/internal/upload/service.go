package upload

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"io"
	"log"
	"net/http"
	"strings"
	"tech.low-stack.temp/server/internal/env"
	"tech.low-stack.temp/server/internal/storage"
)

type LimitWriter struct {
	w       io.Writer
	limit   uint64
	written uint64
}

func (l *LimitWriter) Write(p []byte) (n int, err error) {
	if l.written+uint64(len(p)) > l.limit {
		return 0, fmt.Errorf("file size exceeds limit of %d bytes", l.limit)
	}

	n, err = l.w.Write(p)
	l.written += uint64(n)
	return
}

func Initialize() {
	http.Handle("POST /", http.HandlerFunc(handleUpload))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	// Get a multipart reader
	reader, err := r.MultipartReader()
	if err != nil {
		log.Printf("Unable to get multipart reader:\n%s", err.Error())
		http.Error(w, "Unable to process upload", http.StatusBadRequest)
		return
	}

	// Read form parts
	part, err := reader.NextPart()
	if err != nil {
		log.Printf("Unable to get form part:\n%s", err.Error())
		http.Error(w, "Unable to process upload", http.StatusBadRequest)
		return
	}

	// Check if this is the file part
	if part.FormName() != "file" {
		log.Printf("No file field found in form")
		http.Error(w, "No file field found in form", http.StatusBadRequest)
		return
	}

	// Get filename
	filename := part.FileName()
	if filename == "" {
		log.Printf("No filename provided")
		http.Error(w, "No filename provided", http.StatusBadRequest)
		return
	}

	// Parse expiration duration (you'll need to modify this since we can't use r.FormValue anymore)
	expiration := env.DefaultExpiration // Set default for now

	// Create file in database and get io writer
	file, databaseFile, err := storage.RequestNewFile(filename, expiration, r.Context())
	if err != nil {
		log.Printf("Unable to request upload: %s", err.Error())
		http.Error(w, "Unable to request upload", http.StatusInternalServerError)
		return
	}

	// Calculate write limit
	freeStorageSpace, _ := storage.GetFreeSpace()
	writeLimit := freeStorageSpace - env.MinFreeSpace
	if env.MaxFileSize < writeLimit {
		writeLimit = env.MaxFileSize
	}

	// Stream the file directly to storage
	limitWriter := &LimitWriter{w: file, limit: writeLimit}
	_, err = io.Copy(limitWriter, part)
	if err != nil {
		if strings.Contains(err.Error(), "file size exceeds limit") {
			log.Printf("File too large, exceeds %s", humanize.Bytes(writeLimit))
			http.Error(w, fmt.Sprintf("File too large! Exceeds %s", humanize.Bytes(writeLimit)), http.StatusRequestEntityTooLarge)
			return
		}
		log.Printf("Unable to write to upload: %s", err.Error())
		http.Error(w, "Unable to write to upload", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(databaseFile.GetDownloadURL()))
	log.Printf("Uploaded %s (%s)", databaseFile.Filename, databaseFile.ID)
}
