package group

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"tech.low-stack.temp/server/internal/archivename"
	"tech.low-stack.temp/server/internal/db"
	"tech.low-stack.temp/server/internal/env"
	"tech.low-stack.temp/server/internal/storage"
	"tech.low-stack.temp/shared/http_error"
	"tech.low-stack.temp/shared/time_utils"
)

type createRequest struct {
	Expiration  string `json:"expiration"`
	ArchiveName string `json:"archiveName"`
}

func Initialize() {
	http.Handle("POST /api/groups", http.HandlerFunc(createGroup))
	http.Handle("POST /api/groups/{id}/files", http.HandlerFunc(uploadFile))
	http.Handle("POST /api/groups/{id}/complete", http.HandlerFunc(completeGroup))
	http.Handle("GET /api/groups/{id}", http.HandlerFunc(getGroup))
}

func createGroup(w http.ResponseWriter, r *http.Request) {
	var request createRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&request)
	}
	expiration := env.DefaultExpiration
	var err error
	if request.Expiration != "" {
		expiration, err = time_utils.ParseDuration(request.Expiration)
		if expiration == 0 && err == nil {
			expiration = env.DefaultExpiration
		}
	}
	if err != nil || expiration < env.MinExpiration || expiration > env.MaxExpiration {
		http_error.Respond(w, http.StatusBadRequest, fmt.Sprintf("Expiration must be between %s and %s", env.MinExpiration, env.MaxExpiration))
		return
	}
	id, err := storage.RequestNewGroup(r.Context(), expiration, archivename.Normalize(request.ArchiveName))
	if err != nil {
		http.Error(w, "Unable to create upload", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := storage.GetGroup(r.Context(), id); err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Unable to process upload", http.StatusBadRequest)
		return
	}
	part, err := reader.NextPart()
	if err != nil || part.FormName() != "file" || part.FileName() == "" {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	filename := filepath.Base(strings.TrimSpace(part.FileName()))
	if filename == "." || filename == "" {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	free, _ := storage.GetFreeSpace()
	if free <= env.MinFreeSpace {
		http.Error(w, "Not enough storage space", http.StatusInsufficientStorage)
		return
	}
	limit := free - env.MinFreeSpace
	if env.MaxFileSize < limit {
		limit = env.MaxFileSize
	}
	file, err := storage.StoreGroupFile(r.Context(), id, filename, part, limit)
	if err != nil {
		if strings.Contains(err.Error(), "exceeds limit") {
			http_error.Respond(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("File too large! Exceeds %s", humanize.Bytes(limit)))
		} else {
			http.Error(w, "Unable to store file", http.StatusInternalServerError)
		}
		return
	}
	writeJSON(w, http.StatusCreated, fileResponse(id, file.ID, file.Filename, file.Size))
}

func completeGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	group, err := storage.GetGroup(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	files, err := storage.GetGroupFiles(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}
	if err := storage.FinalizeGroup(r.Context(), id); err != nil {
		http.Error(w, "Unable to finalize upload", http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-Download-Page", fmt.Sprintf("%s/d/%s", env.BaseUrl, id))
	writeJSON(w, http.StatusOK, groupResponse(id, files, group.ExpiresAt, group.ArchiveName))
}

func getGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	group, err := storage.GetGroup(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	files, err := storage.GetGroupFiles(r.Context(), id)
	if err != nil {
		http.Error(w, "Unable to read upload", http.StatusInternalServerError)
		return
	}
	if !group.Finalized {
		http.Error(w, "Upload is not finalized", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, groupResponse(id, files, group.ExpiresAt, group.ArchiveName))
}

type fileJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	URL  string `json:"url"`
}
type groupJSON struct {
	ID        string     `json:"id"`
	DirectURL string     `json:"directUrl"`
	PageURL   string     `json:"pageUrl"`
	ExpiresAt string     `json:"expiresAt"`
	Files     []fileJSON `json:"files"`
}

func fileResponse(groupID, id, name string, size int64) fileJSON {
	return fileJSON{ID: id, Name: name, Size: size, URL: fmt.Sprintf("%s/g/%s/%s/%s", env.BaseUrl, groupID, id, name)}
}
func groupResponse(id string, files []db.GroupFile, expiresAt time.Time, archiveName *string) groupJSON {
	result := groupJSON{
		ID:        id,
		PageURL:   fmt.Sprintf("%s/d/%s", env.BaseUrl, id),
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
		Files:     make([]fileJSON, 0, len(files)),
	}
	for _, file := range files {
		result.Files = append(result.Files, fileResponse(id, file.ID, file.Filename, file.Size))
	}
	if len(files) == 1 {
		result.DirectURL = result.Files[0].URL
	} else {
		result.DirectURL = fmt.Sprintf("%s/f/%s/%s", env.BaseUrl, id, url.PathEscape(archivename.Filename(archiveName, files)))
	}
	return result
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
