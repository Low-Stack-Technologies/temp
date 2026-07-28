package archivename

import (
	"path/filepath"
	"strings"
	"unicode"

	"tech.low-stack.temp/server/internal/db"
)

func Normalize(custom string) *string {
	name := normalizeStem(custom)
	if name == "" {
		return nil
	}
	return &name
}

func Filename(custom *string, files []db.GroupFile) string {
	if custom != nil {
		return *custom + ".zip"
	}
	stems := make([]string, 0, len(files))
	for _, file := range files {
		stems = append(stems, normalizeStem(file.Filename))
	}
	if len(stems) == 0 {
		return "files.zip"
	}
	prefix := stems[0]
	for _, stem := range stems[1:] {
		for !strings.HasPrefix(stem, prefix) && prefix != "" {
			prefix = prefix[:len(prefix)-1]
		}
	}
	prefix = strings.TrimRight(prefix, " .-_()[]")
	if prefix == "" {
		return "files.zip"
	}
	return prefix + ".zip"
}

func normalizeStem(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = filepath.Base(value)
	if ext := filepath.Ext(value); ext != "" {
		value = strings.TrimSuffix(value, ext)
	}
	var builder strings.Builder
	for _, char := range value {
		if unicode.IsControl(char) {
			continue
		}
		builder.WriteRune(char)
	}
	result := strings.TrimSpace(builder.String())
	if result == "." || result == ".." {
		return ""
	}
	return result
}
