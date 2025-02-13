package storage

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"time"

	"github.com/ricochet2200/go-disk-usage/du"
	"tech.low-stack.temp/server/internal/db"
	"tech.low-stack.temp/server/internal/env"
)

func RequestNewFile(ctx context.Context) (io.WriteCloser, *db.File, error) {
	id := newUuid()
	qtx := db.NewQueries()

	databaseFile, err := qtx.CreateFile(ctx, db.CreateFileParams{
		ID:         id,
		Expiration: int64(10),
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

func UpdateFile(id string, filename string, expiration time.Duration, ctx context.Context) (*db.File, error) {
	qtx := db.NewQueries()

	databaseFile, err := qtx.UpdateFile(ctx, db.UpdateFileParams{
		ID:         id,
		Filename:   &filename,
		Expiration: int64(math.Round(expiration.Minutes())),
	})

	return &databaseFile, err
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

func GetFreeSpace() (uint64, error) {
	diskUsage := du.NewDiskUsage(env.StoragePath)
	if diskUsage == nil {
		return 0, fmt.Errorf("failed to get disk usage")
	}

	return diskUsage.Free(), nil
}

func GetStoragePath(id string) string {
	return path.Join(env.StoragePath, id)
}
