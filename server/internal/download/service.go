package download

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"tech.low-stack.temp/server/internal/archivename"
	"tech.low-stack.temp/server/internal/storage"
)

func Initialize() {
	http.Handle("GET /f/", http.HandlerFunc(handleDownload))
	http.Handle("GET /g/", http.HandlerFunc(handleGroupFile))
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/f/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id := parts[0]
	if group, err := storage.GetGroup(r.Context(), id); err == nil {
		if !group.Finalized {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		serveZip(w, r, id)
		return
	}
	file, databaseFile, err := storage.GetFile(id, r.Context())
	if err != nil || databaseFile == nil || file == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileStat, err := os.Stat(storage.GetStoragePath(databaseFile.ID))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fileName := "file"
	if databaseFile.Filename != nil {
		fileName = *databaseFile.Filename
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileStat.Size()))
	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, file)
}

func handleGroupFile(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/g/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	group, err := storage.GetGroup(r.Context(), parts[0])
	if err != nil || !group.Finalized {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	groupFile, err := storage.GetGroupFile(r.Context(), parts[0], parts[1])
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	file, err := os.Open(storage.GetStoragePath(groupFile.ID))
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(groupFile.Filename)))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", groupFile.Size))
	_, _ = io.Copy(w, file)
}

func serveZip(w http.ResponseWriter, r *http.Request, groupID string) {
	group, err := storage.GetGroup(r.Context(), groupID)
	if err != nil || !group.Finalized {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	files, err := storage.GetGroupFiles(r.Context(), groupID)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", archivename.Filename(group.ArchiveName, files)))
	w.Header().Set("Content-Type", "application/zip")
	archive := zip.NewWriter(w)
	usedNames := make(map[string]int)
	for _, databaseFile := range files {
		name := filepath.Base(databaseFile.Filename)
		usedNames[name]++
		if usedNames[name] > 1 {
			ext := filepath.Ext(name)
			base := strings.TrimSuffix(name, ext)
			name = fmt.Sprintf("%s (%d)%s", base, usedNames[databaseFile.Filename], ext)
		}
		entry, err := archive.Create(name)
		if err != nil {
			_ = archive.Close()
			return
		}
		file, err := os.Open(storage.GetStoragePath(databaseFile.ID))
		if err != nil {
			_ = archive.Close()
			return
		}
		_, copyErr := io.Copy(entry, file)
		file.Close()
		if copyErr != nil {
			_ = archive.Close()
			return
		}
	}
	_ = archive.Close()
}
