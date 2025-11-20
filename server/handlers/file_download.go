package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Low-Stack-Technologies/temp/server/db"
	"github.com/Low-Stack-Technologies/temp/server/storage"
)

// FileDownloadHandler handles file downloads
type FileDownloadHandler struct {
	repo    *db.Repository
	storage *storage.Storage
}

// NewFileDownloadHandler creates a new file download handler
func NewFileDownloadHandler(repo *db.Repository, stor *storage.Storage) *FileDownloadHandler {
	return &FileDownloadHandler{
		repo:    repo,
		storage: stor,
	}
}

// ServeHTTP handles the download request
func (h *FileDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract file ID from path
	fileID := r.URL.Path[len("/api/download/"):]
	if fileID == "" {
		http.Error(w, "file ID required", http.StatusBadRequest)
		return
	}
	
	// Get file record from database
	record, err := h.repo.GetFileByID(fileID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	
	if record == nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	
	// Check if file has expired
	if time.Now().After(record.ExpiresAt) {
		http.Error(w, "file has expired", http.StatusNotFound)
		return
	}
	
	// Open file
	filePath := h.storage.GetFilePath(record.StoragePath)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()
	
	// Set headers
	w.Header().Set("Content-Type", record.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", record.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", record.Size))
	w.Header().Set("X-Checksum-SHA256", record.Checksum)
	
	// Stream file content
	if _, err := io.Copy(w, file); err != nil {
		// Can't send error response after headers are sent
		// Just log it
		return
	}
}
