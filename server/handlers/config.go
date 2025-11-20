package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Low-Stack-Technologies/temp/server/config"
)

// ConfigHandler serves the public configuration
type ConfigHandler struct {
	cfg *config.Config
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{cfg: cfg}
}

// ConfigResponse represents the public configuration
type ConfigResponse struct {
	MaxUploadSizeMB   int64 `json:"max_upload_size_mb"`
	MaxFilesPerUpload int   `json:"max_files_per_upload"`
	MinTTLSeconds     int   `json:"min_ttl_seconds"`
	MaxTTLSeconds     int   `json:"max_ttl_seconds"`
}

// ServeHTTP handles the config request
func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := ConfigResponse{
		MaxUploadSizeMB:   h.cfg.MaxUploadSizeMB,
		MaxFilesPerUpload: h.cfg.MaxFilesPerUpload,
		MinTTLSeconds:     h.cfg.MinTTLSeconds,
		MaxTTLSeconds:     h.cfg.MaxTTLSeconds,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
