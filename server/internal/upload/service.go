package upload

import (
	"fmt"
	"io"
	"net/http"
	"tech.low-stack.temp/server/internal/env"
	"tech.low-stack.temp/server/internal/storage"
)

func Initialize() {
	http.Handle("POST /", http.HandlerFunc(handleUpload))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse multipart form:\n%s", err.Error()), http.StatusBadRequest)
		return
	}

	uploadedFile, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to get file:\n%s", err.Error()), http.StatusBadRequest)
		return
	}
	defer uploadedFile.Close()

	file, databaseFile, err := storage.RequestNewFile(header.Filename, env.DefaultExpiration, r.Context())
	if err != nil {
		http.Error(w, "Unable to request upload", http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(file, uploadedFile)
	if err != nil {
		http.Error(w, "Unable to write to upload", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(databaseFile.GetDownloadURL()))
}
