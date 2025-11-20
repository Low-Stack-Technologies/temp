package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Low-Stack-Technologies/temp/server/config"
	"github.com/Low-Stack-Technologies/temp/server/db"
)

// FileListHandler handles file listing requests
type FileListHandler struct {
	cfg  *config.Config
	repo *db.Repository
}

// NewFileListHandler creates a new file list handler
func NewFileListHandler(cfg *config.Config, repo *db.Repository) *FileListHandler {
	return &FileListHandler{
		cfg:  cfg,
		repo: repo,
	}
}

// ListResponse represents the file list response
type ListResponse struct {
	Files []FileMetadata `json:"files"`
	Total int            `json:"total"`
}

// ServeHTTP handles the file list request
func (h *FileListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 50
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	// Get files from database
	records, err := h.repo.ListFiles(limit, offset)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	
	// Convert to response format
	files := make([]FileMetadata, 0, len(records))
	for _, record := range records {
		files = append(files, FileMetadata{
			ID:          record.ID,
			Filename:    record.Filename,
			Size:        record.Size,
			MimeType:    record.MimeType,
			DownloadURL: h.cfg.BaseURL + "/api/download/" + record.ID,
			UploadedAt:  record.UploadedAt,
			ExpiresAt:   record.ExpiresAt,
			Checksum:    record.Checksum,
		})
	}
	
	response := ListResponse{
		Files: files,
		Total: len(files),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// FileMetadataHandler handles single file metadata requests
type FileMetadataHandler struct {
	cfg  *config.Config
	repo *db.Repository
}

// NewFileMetadataHandler creates a new file metadata handler
func NewFileMetadataHandler(cfg *config.Config, repo *db.Repository) *FileMetadataHandler {
	return &FileMetadataHandler{
		cfg:  cfg,
		repo: repo,
	}
}

// ServeHTTP handles the file metadata request
func (h *FileMetadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract file ID from path
	fileID := r.URL.Path[len("/api/files/"):]
	if fileID == "" {
		http.Error(w, "file ID required", http.StatusBadRequest)
		return
	}
	
	// Get file record
	record, err := h.repo.GetFileByID(fileID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	
	if record == nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	
	// Convert to response format
	metadata := FileMetadata{
		ID:          record.ID,
		Filename:    record.Filename,
		Size:        record.Size,
		MimeType:    record.MimeType,
		DownloadURL: h.cfg.BaseURL + "/api/download/" + record.ID,
		UploadedAt:  record.UploadedAt,
		ExpiresAt:   record.ExpiresAt,
		Checksum:    record.Checksum,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}
