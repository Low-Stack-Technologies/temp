package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"tech.low-stack.temp/cli/internal/env"
	"tech.low-stack.temp/cli/internal/update"
	"tech.low-stack.temp/cli/internal/upload"
	"tech.low-stack.temp/shared/time_utils"
)

func main() {
	env.LoadVariables()
	update.CheckVersion()

	expirationStr := pflag.StringP("expiration", "e", "", "Set expiration time (e.g., 5h)")
	pflag.Parse()

	expiration, err := time_utils.ParseDuration(*expirationStr)
	if err != nil && expirationStr != nil {
		fmt.Printf("invalid argument \"%s\" for \"-e, --expiration\" flag:\n%s\n", *expirationStr, err.Error())
		os.Exit(1)
	}

	if expirationStr == nil {
		expiration = time.Duration(0)
	}

	filePaths := pflag.Args()
	uploadFilesIndividually(filePaths, expiration)
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
