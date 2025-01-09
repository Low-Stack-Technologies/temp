package storage

import (
	"context"
	"io"
	"os"
	"path"
	"tech.low-stack.temp/server/internal/db"
	"tech.low-stack.temp/server/internal/env"
)

func RequestNewFile(filename string, ctx context.Context) (io.WriteCloser, *db.File, error) {
	id := newUuid()
	qtx := db.NewQueries()

	databaseFile, err := qtx.CreateFile(ctx, db.CreateFileParams{ID: id, Filename: filename})
	if err != nil {
		return nil, nil, err
	}

	filePath := GetStoragePath(id)
	fileWriter, err := os.Create(filePath)
	if err != nil {
		return nil, nil, err
	}

	return fileWriter, &databaseFile, nil
}

func GetFile(id string) (io.ReadCloser, *db.File, error) {
	qtx := db.NewQueries()

	databaseFile, err := qtx.GetFile(context.Background(), id)
	if err != nil {
		return nil, nil, err
	}

	filePath := GetStoragePath(id)
	fileReader, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}

	return fileReader, &databaseFile, nil
}

func GetStoragePath(id string) string {
	return path.Join(env.StoragePath, id)
}
