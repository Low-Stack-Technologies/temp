package upload

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"tech.low-stack.temp/cli/internal/env"
	"tech.low-stack.temp/shared/http_error"
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

	// Start upload in goroutine
	errChan := make(chan error)
	go func() {
		var pipeErr error
		defer func() {
			writer.Close()
			if pipeErr != nil {
				pw.CloseWithError(pipeErr)
			} else {
				pw.Close()
			}
			close(errChan)
		}()

		// Create form file field
		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			pipeErr = fmt.Errorf("failed to create form file: %w", err)
			return
		}

		// Copy file to multipart writer through progress reader
		if _, err := io.Copy(part, progress); err != nil {
			pipeErr = fmt.Errorf("failed to copy file: %w", err)
			return
		}

		// Add expiration field to form
		if err := writer.WriteField("expiration", expiration.String()); err != nil {
			pipeErr = fmt.Errorf("failed to add expiration field: %w", err)
			return
		}
	}()

	// Send the request
	resp, err := sendRequest(pr, writer)
	if err != nil {
		return "", err
	}

	// Wait for any errors from the goroutine
	if err := <-errChan; err != nil {
		return "", err
	}

	return resp, nil
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)

	progressMutex.Lock()
	pr.BytesRead += int64(n)
	pr.Percentage = float64(pr.BytesRead) / float64(pr.Size) * 100
	progressMutex.Unlock()

	return n, err
}

func sendRequest(pr *io.PipeReader, writer *multipart.Writer) (string, error) {
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

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		// Attempt to parse error message
		errorMessage := http_error.Error{}
		err = json.Unmarshal(respBody, &errorMessage)
		if err != nil {
			return "", fmt.Errorf("failed to parse error message: %w", err)
		}

		fmt.Printf("Error: %s\n", errorMessage.Message)
		return "", fmt.Errorf("failed to upload file")
	}

	return string(respBody), nil
}
