package web

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"tech.low-stack.temp/server/internal/env"
)

//go:embed public
var publicDir embed.FS

func Initialize() {
	fsys, err := fs.Sub(publicDir, "public")
	if err != nil {
		log.Fatalf("Failed to initialize web server: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			serveIndexHTML(w, fsys)
			return
		}

		http.FileServer(http.FS(fsys)).ServeHTTP(w, r)
	})
}

func serveIndexHTML(w http.ResponseWriter, fsys fs.FS) {
	file, err := fsys.Open("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Replace placeholders with actual values
	contents = bytes.ReplaceAll(contents, []byte("/*MINIMUM_EXPIRATION_TIME_SECONDS*/"), []byte(strconv.Itoa(int(env.MinExpiration.Seconds()))))
	contents = bytes.ReplaceAll(contents, []byte("/*MAXIMUM_EXPIRATION_TIME_SECONDS*/"), []byte(strconv.Itoa(int(env.MaxExpiration.Seconds()))))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(contents)
}
