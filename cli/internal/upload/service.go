package upload

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"tech.low-stack.temp/cli/internal/env"
	"time"
)

type ProgressReader struct {
	Filename   string
	Index      int
	Reader     io.Reader
	Size       int64
	BytesRead  int64
	Percentage float64
}

func UploadFile(filePath string, index int, expiration time.Duration) (string, error) {
	// Open file to upload
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	// Get file stats for progress calculation
	fileStats, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("unable to read file stats: %w", err)
	}
	fileSize := fileStats.Size()

	// Create pipe for streaming contents
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// Create progress reader
	progress := &ProgressReader{
		Filename: filepath.Base(filePath),
		Index:    index,
		Reader:   file,
		Size:     fileSize,
	}

	// Add progress reader to progress readers
	progressBars = append(progressBars, progress)

	// Write the multipart form in a goroutine
	go func() {
		defer pw.Close()
		defer writer.Close()

		// Create form file field
		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to create form file: %w", err))
			return
		}

		// Copy file to multipart writer through progress reader
		if _, err := io.Copy(part, progress); err != nil {
			pw.CloseWithError(fmt.Errorf("failed to copy file: %w", err))
			return
		}

		// Add expiration field to form
		if err := writer.WriteField("expiration", expiration.String()); err != nil {
			pw.CloseWithError(fmt.Errorf("failed to add expiration field: %w", err))
			return
		}
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
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(respBody), nil
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)

	progressMutex.Lock()
	pr.BytesRead += int64(n)
	pr.Percentage = float64(pr.BytesRead) / float64(pr.Size) * 100
	progressMutex.Unlock()

	return n, err
}
