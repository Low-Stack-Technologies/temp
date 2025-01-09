package db

import (
	"fmt"
	"tech.low-stack.temp/server/internal/env"
)

func (f *File) GetDownloadURL() string {
	return fmt.Sprintf("%s/f/%s/%s", env.BaseUrl, f.ID, f.Filename)
}
