package upload

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"tech.low-stack.temp/cli/internal/env"
)

type ProgressReader struct {
	Filename   string
	Index      int
	Reader     io.Reader
	Size       int64
	BytesRead  int64
	Percentage float64
}

func UploadFile(filePath string, index int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %w", err)
	}

	// Create pipe to monitor progress
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// Get file stats for progress calculation
	fileStats, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("unable to read file stats: %w", err)
	}
	fileSize := fileStats.Size()

	go func() {
		defer pw.Close()

		// Create form field
		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			return
		}

		// Create progress reader
		progress := &ProgressReader{
			Filename: filepath.Base(filePath),
			Index:    index,
			Reader:   file,
			Size:     fileSize,
		}

		// Add progress reader to progress readers
		progressBars = append(progressBars, progress)

		// Copy file to multipart writer through progress reader
		io.Copy(part, progress)
		writer.Close()
	}()

	// Create request
	req, err := http.NewRequest("POST", env.ServiceUrl, pr)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)

	progressMutex.Lock()
	pr.BytesRead += int64(n)
	pr.Percentage = float64(pr.BytesRead) / float64(pr.Size) * 100
	progressMutex.Unlock()

	return n, err
}
