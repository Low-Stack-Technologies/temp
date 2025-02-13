package db

import (
	"fmt"

	"tech.low-stack.temp/server/internal/env"
)

func (f *File) GetDownloadURL() string {
	if f.Filename == nil {
		return fmt.Sprintf("%s/f/%s/file", env.BaseUrl, f.ID)
	}
	return fmt.Sprintf("%s/f/%s/%s", env.BaseUrl, f.ID, *f.Filename)
}
