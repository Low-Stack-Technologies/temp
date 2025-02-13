package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"tech.low-stack.temp/cli/internal/env"
	"tech.low-stack.temp/cli/internal/update"
	"tech.low-stack.temp/cli/internal/upload"
)

func main() {
	env.LoadVariables()
	update.CheckVersion()

	expiration := pflag.DurationP("expiration", "e", time.Duration(0), "Set expiration time (e.g., 5h)")
	pflag.Parse()

	filePaths := pflag.Args()
	uploadFilesIndividually(filePaths, *expiration)
}

func uploadFilesIndividually(filePaths []string, expiration time.Duration) {
	var wg sync.WaitGroup
	downloadUrls := make([]string, len(filePaths))

	for i, filePath := range filePaths {
		fmt.Println()
		wg.Add(1)
		go func(path string, index int) {
			defer wg.Done()
			downloadUrl, err := upload.UploadFile(path, index, expiration)
			if err != nil {
				fmt.Printf("failed to upload file: %s\n", err)
			}

			downloadUrls[index] = downloadUrl
		}(filePath, i)
	}

	go func() {
		for {
			upload.DrawAllProgressBars()
			time.Sleep(time.Millisecond * 100)
		}
	}()

	wg.Wait()

	for _, downloadUrl := range downloadUrls {
		fmt.Println(downloadUrl)
	}
}
