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

	// Attempt to delete file from storage
	err := os.Remove(GetStoragePath(id))
	if err != nil && !os.IsNotExist(err) {
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

func RequestNewGroup(ctx context.Context, expiration time.Duration, archiveName *string) (string, error) {
	id := newUuid()
	_, err := db.CreateGroup(ctx, id, expiration, archiveName)
	return id, err
}

func StoreGroupFile(ctx context.Context, groupID, filename string, src io.Reader, limit uint64) (db.GroupFile, error) {
	id := newUuid()
	filePath := GetStoragePath(id)
	fileWriter, err := os.Create(filePath)
	if err != nil {
		return db.GroupFile{}, err
	}
	written, err := io.Copy(fileWriter, io.LimitReader(src, int64(limit)+1))
	if err != nil {
		fileWriter.Close()
		_ = os.Remove(filePath)
		return db.GroupFile{}, err
	}
	if written > int64(limit) {
		fileWriter.Close()
		_ = os.Remove(filePath)
		return db.GroupFile{}, fmt.Errorf("file size exceeds limit")
	}
	if err = fileWriter.Close(); err != nil {
		_ = os.Remove(filePath)
		return db.GroupFile{}, err
	}
	file := db.GroupFile{ID: id, GroupID: groupID, Filename: filename, Size: written}
	if err = db.CreateGroupFile(ctx, file); err != nil {
		_ = os.Remove(filePath)
		return db.GroupFile{}, err
	}
	return file, nil
}

func GetGroup(ctx context.Context, id string) (db.Group, error) { return db.GetGroup(ctx, id) }

func GetGroupFiles(ctx context.Context, id string) ([]db.GroupFile, error) {
	return db.GetGroupFiles(ctx, id)
}

func GetGroupFile(ctx context.Context, groupID, fileID string) (db.GroupFile, error) {
	return db.GetGroupFile(ctx, groupID, fileID)
}

func FinalizeGroup(ctx context.Context, id string) error { return db.FinalizeGroup(ctx, id) }

func DeleteGroup(ctx context.Context, id string) error {
	files, err := db.GetGroupFiles(ctx, id)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := os.Remove(GetStoragePath(file.ID)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return db.DeleteGroup(ctx, id)
}

func GetExpiredGroups(ctx context.Context) ([]db.Group, error) { return db.GetExpiredGroups(ctx) }
