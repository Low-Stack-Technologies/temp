package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run http_client.go <file1> [file2] [file3] ...")
		os.Exit(1)
	}
	
	// Configuration
	serverURL := "http://localhost:8080/api/upload"
	ttlSeconds := 3600 // 1 hour
	
	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// Add TTL parameter
	writer.WriteField("ttl_seconds", fmt.Sprintf("%d", ttlSeconds))
	
	// Add files
	for _, filePath := range os.Args[1:] {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", filePath, err)
			continue
		}
		defer file.Close()
		
		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			fmt.Printf("Error creating form file: %v\n", err)
			continue
		}
		
		if _, err := io.Copy(part, file); err != nil {
			fmt.Printf("Error copying file: %v\n", err)
			continue
		}
		
		fmt.Printf("Added file: %s\n", filePath)
	}
	
	writer.Close()
	
	// Create request
	req, err := http.NewRequest("POST", serverURL, &buf)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	// Send request
	fmt.Println("\nUploading files...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("\nStatus: %s\n", resp.Status)
	fmt.Printf("Response:\n%s\n", string(body))
}
