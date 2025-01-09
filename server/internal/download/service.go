package download

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"tech.low-stack.temp/server/internal/storage"
)

func Initialize() {
	http.Handle("GET /f/", http.HandlerFunc(handleDownload))
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/f/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id := parts[0]
	file, databaseFile, err := storage.GetFile(id, r.Context())
	if err != nil || databaseFile == nil || file == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", databaseFile.Filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, file)
}
