package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tech.low-stack.temp/cli/internal/env"
)

type groupResponse struct {
	DirectURL string `json:"directUrl"`
	PageURL   string `json:"pageUrl"`
}
type createGroupResponse struct {
	ID string `json:"id"`
}

func UploadGroup(filePaths []string, expiration time.Duration, archiveName string) (groupResponse, error) {
	progressMutex.Lock()
	progressBars = nil
	progressMutex.Unlock()
	var created createGroupResponse
	expirationValue := ""
	if expiration != 0 {
		expirationValue = expiration.String()
	}
	body, _ := json.Marshal(map[string]string{"expiration": expirationValue, "archiveName": archiveName})
	request, err := http.NewRequest(http.MethodPost, env.ServiceUrl+"/api/groups", bytes.NewReader(body))
	if err != nil {
		return groupResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{}).Do(request)
	if err != nil {
		return groupResponse{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return groupResponse{}, fmt.Errorf("server does not support grouped uploads")
	}
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		return groupResponse{}, err
	}
	if created.ID == "" {
		return groupResponse{}, fmt.Errorf("server returned an invalid upload id")
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(filePaths))
	for i, filePath := range filePaths {
		wg.Add(1)
		go func(path string, index int) {
			defer wg.Done()
			if err := uploadGroupFile(path, created.ID, index); err != nil {
				errs <- err
			}
		}(filePath, i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		return groupResponse{}, err
	}

	request, err = http.NewRequest(http.MethodPost, env.ServiceUrl+"/api/groups/"+created.ID+"/complete", nil)
	if err != nil {
		return groupResponse{}, err
	}
	response, err = (&http.Client{}).Do(request)
	if err != nil {
		return groupResponse{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return groupResponse{}, fmt.Errorf("unable to finalize grouped upload")
	}
	var result groupResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return groupResponse{}, err
	}
	return result, nil
}

func uploadGroupFile(filePath, groupID string, index int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	stats, err := file.Stat()
	if err != nil {
		return err
	}
	progress := &ProgressReader{Filename: filepath.Base(filePath), Index: index, Reader: file, Size: stats.Size()}
	progressMutex.Lock()
	progressBars = append(progressBars, progress)
	progressMutex.Unlock()
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	errChan := make(chan error, 1)
	go func() {
		part, e := writer.CreateFormFile("file", filepath.Base(filePath))
		if e == nil {
			_, e = io.Copy(part, progress)
		}
		if closeErr := writer.Close(); e == nil {
			e = closeErr
		}
		_ = pw.Close()
		errChan <- e
		close(errChan)
	}()
	request, err := http.NewRequest(http.MethodPost, env.ServiceUrl+"/api/groups/"+groupID+"/files", pr)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := (&http.Client{}).Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err := <-errChan; err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("failed to upload %s", filepath.Base(filePath))
	}
	return nil
}
