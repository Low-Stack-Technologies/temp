package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Low-Stack-Technologies/temp/server/config"
	"github.com/Low-Stack-Technologies/temp/server/db"
	"github.com/Low-Stack-Technologies/temp/server/storage"
	"github.com/google/uuid"
)

// HTTPUploadHandler handles HTTP file uploads
type HTTPUploadHandler struct {
	cfg     *config.Config
	repo    *db.Repository
	storage *storage.Storage
}

// NewHTTPUploadHandler creates a new HTTP upload handler
func NewHTTPUploadHandler(cfg *config.Config, repo *db.Repository, stor *storage.Storage) *HTTPUploadHandler {
	return &HTTPUploadHandler{
		cfg:     cfg,
		repo:    repo,
		storage: stor,
	}
}

// UploadResponse represents the JSON response for uploads
type UploadResponse struct {
	Success   bool           `json:"success"`
	Files     []FileMetadata `json:"files,omitempty"`
	ExpiresAt time.Time      `json:"expires_at,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// FileMetadata represents file information in responses
type FileMetadata struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	DownloadPath string    `json:"download_path"`
	UploadedAt   time.Time `json:"uploaded_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Checksum     string    `json:"checksum"`
}

// ServeHTTP handles the HTTP upload request
func (h *HTTPUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	
	// Parse multipart form with size limit
	maxSize := h.cfg.MaxUploadSizeBytes()
	if err := r.ParseMultipartForm(maxSize); err != nil {
		h.sendError(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}
	defer r.MultipartForm.RemoveAll()
	
	// Get TTL parameter
	ttlStr := r.FormValue("ttl_seconds")
	if ttlStr == "" {
		h.sendError(w, http.StatusBadRequest, "ttl_seconds is required")
		return
	}
	
	ttlSeconds, err := strconv.Atoi(ttlStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid ttl_seconds")
		return
	}
	
	// Validate TTL
	if ttlSeconds < h.cfg.MinTTLSeconds || ttlSeconds > h.cfg.MaxTTLSeconds {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("ttl_seconds must be between %d and %d", h.cfg.MinTTLSeconds, h.cfg.MaxTTLSeconds))
		return
	}
	
	// Get uploaded files
	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		h.sendError(w, http.StatusBadRequest, "no files uploaded")
		return
	}
	
	if len(files) > h.cfg.MaxFilesPerUpload {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("too many files (max %d)", h.cfg.MaxFilesPerUpload))
		return
	}
	
	// Calculate expiration time
	uploadedAt := time.Now()
	expiresAt := uploadedAt.Add(time.Duration(ttlSeconds) * time.Second)
	
	var uploadedFiles []FileMetadata
	
	// Process each file
	for _, fileHeader := range files {
		// Check file size
		if fileHeader.Size > maxSize {
			h.sendError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("file %s exceeds maximum size", fileHeader.Filename))
			return
		}
		
		// Open uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, "failed to open uploaded file")
			return
		}
		
		// Save file
		storagePath, checksum, size, err := h.storage.SaveFile(file, fileHeader.Filename)
		file.Close()
		
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, "failed to save file: "+err.Error())
			return
		}
		
		// Detect MIME type
		mimeType := fileHeader.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		
		// Create database record
		fileID := uuid.New().String()
		record := &db.FileRecord{
			ID:           fileID,
			Filename:     fileHeader.Filename,
			Size:         size,
			MimeType:     mimeType,
			StoragePath:  storagePath,
			UploadMethod: "http",
			UploadedAt:   uploadedAt,
			TTLSeconds:   ttlSeconds,
			ExpiresAt:    expiresAt,
			Checksum:     checksum,
		}
		
		if err := h.repo.CreateFileRecord(record); err != nil {
			// Clean up file on database error
			h.storage.DeleteFile(storagePath)
			h.sendError(w, http.StatusInternalServerError, "failed to save file metadata")
			return
		}
		
		// Add to response
		downloadPath := fmt.Sprintf("/api/download/%s", fileID)
		uploadedFiles = append(uploadedFiles, FileMetadata{
			ID:           fileID,
			Filename:     fileHeader.Filename,
			Size:         size,
			MimeType:     mimeType,
			DownloadPath: downloadPath,
			UploadedAt:   uploadedAt,
			ExpiresAt:    expiresAt,
			Checksum:     checksum,
		})
	}
	
	// Send success response
	response := UploadResponse{
		Success:   true,
		Files:     uploadedFiles,
		ExpiresAt: expiresAt,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendError sends a JSON error response
func (h *HTTPUploadHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	response := UploadResponse{
		Success: false,
		Error:   message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
