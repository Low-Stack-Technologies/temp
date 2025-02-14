package upload

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"tech.low-stack.temp/server/internal/env"
	"tech.low-stack.temp/server/internal/storage"
	"tech.low-stack.temp/shared/http_error"
	"tech.low-stack.temp/shared/time_utils"
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
	reader, err := r.MultipartReader()
	if err != nil {
		log.Printf("Unable to get multipart reader:\n%s", err.Error())
		http.Error(w, "Unable to process upload", http.StatusBadRequest)
		return
	}

	var filename string
	var expiration time.Duration = env.DefaultExpiration
	var written int64

	// Create file in database and get io writer
	file, databaseFile, err := storage.RequestNewFile(r.Context())
	if err != nil {
		log.Printf("Unable to request upload: %s", err.Error())
		http_error.Respond(w, http.StatusInternalServerError, "Unable to request upload! Unable to request new file in database")
		return
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Unable to get form part:\n%s", err.Error())
			http_error.Respond(w, http.StatusBadRequest, "Unable to process upload!\nWas not able to get form part")
			return
		}

		switch part.FormName() {
		case "expiration":
			expirationStr, err := io.ReadAll(part)
			if err != nil {
				log.Printf("Unable to read expiration field: %s", err.Error())
				http_error.Respond(w, http.StatusBadRequest, "Unable to process upload!\nUnable to read expiration field")
				return
			}

			expiration, err = time_utils.ParseDuration(string(expirationStr))
			if err != nil {
				log.Printf("Unable to parse expiration field: %s", err.Error())
				http_error.Respond(w, http.StatusBadRequest, "Unable to process upload!\nUnable to parse expiration field")
				return
			}

			// Set default expiration if none was provided
			if expiration == 0 {
				expiration = env.DefaultExpiration
			}
		case "file":
			// Store filename
			filename = part.FileName()
			if filename == "" {
				log.Printf("No filename provided")
				http_error.Respond(w, http.StatusBadRequest, "No filename provided for file!")
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
			written, err = io.Copy(limitWriter, part)
			if err != nil {
				// Delete file if it was created
				if file != nil {
					storage.DeleteFile(databaseFile.ID, r.Context())
				}

				// Handle file size error
				if strings.Contains(err.Error(), "file size exceeds limit") {
					log.Printf("File too large, exceeds %s", humanize.Bytes(writeLimit))
					http_error.Respond(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("File too large! Exceeds %s", humanize.Bytes(writeLimit)))
					return
				}

				// Handle other errors
				log.Printf("Unable to write to upload: %s", err.Error())
				http_error.Respond(w, http.StatusInternalServerError, "Unable to write to storage! Unknown error")
				return
			}
		}
	}

	// Ensure expiration is within bounds
	if expiration < env.MinExpiration || expiration > env.MaxExpiration {
		log.Printf("Expiration out of bounds: %s", expiration.String())
		http_error.Respond(w, http.StatusBadRequest, fmt.Sprintf("Expiration out of bounds! Must be between %s and %s", env.MinExpiration.String(), env.MaxExpiration.String()))
		return
	}

	// Update file in database
	databaseFile, err = storage.UpdateFile(databaseFile.ID, filename, expiration, r.Context())
	if err != nil {
		log.Printf("Unable to update file: %s", err.Error())
		http_error.Respond(w, http.StatusInternalServerError, "Unable to update file in database!")
		return
	}

	// Respond with download URL
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(databaseFile.GetDownloadURL()))
	log.Printf("Uploaded %s (%s)\t%s", *databaseFile.Filename, databaseFile.ID, humanize.Bytes(uint64(written)))
}
