package storage

import (
	"context"
	"io"
	"math"
	"os"
	"path"
	"tech.low-stack.temp/server/internal/db"
	"tech.low-stack.temp/server/internal/env"
	"time"
)

func RequestNewFile(filename string, expiration time.Duration, ctx context.Context) (io.WriteCloser, *db.File, error) {
	id := newUuid()
	qtx := db.NewQueries()

	databaseFile, err := qtx.CreateFile(ctx, db.CreateFileParams{
		ID:         id,
		Filename:   filename,
		Expiration: int64(math.Round(expiration.Minutes())),
	})
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

func GetFile(id string, ctx context.Context) (io.ReadCloser, *db.File, error) {
	qtx := db.NewQueries()

	databaseFile, err := qtx.GetFile(ctx, id)
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

func DeleteFile(id string, ctx context.Context) error {
	qtx := db.NewQueries()

	if err := os.Remove(GetStoragePath(id)); err != nil {
		return err
	}

	return qtx.DeleteFile(ctx, id)
}

func GetStoragePath(id string) string {
	return path.Join(env.StoragePath, id)
}
