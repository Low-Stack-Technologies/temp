package expiration

import (
	"context"
	"log"
	"time"

	"tech.low-stack.temp/server/internal/db"
	"tech.low-stack.temp/server/internal/storage"
)

func Initialize() {
	go func() {
		for {
			if err := expireFiles(); err != nil {
				log.Printf("Error expiring files: %s\n", err.Error())
			}
			time.Sleep(time.Minute)
		}
	}()
}

func expireFiles() error {
	ctx := context.Background()
	qtx := db.NewQueries()

	expiredFiles, err := qtx.GetExpiredFiles(ctx)
	if err != nil {
		return err
	}

	for _, file := range expiredFiles {
		if err := storage.DeleteFile(file.ID, ctx); err != nil {
			return err
		}

		if file.Filename != nil {
			log.Printf("Expired %s (%s)", *file.Filename, file.ID)
		} else {
			log.Printf("Expired (%s)", file.ID)
		}
	}

	return nil
}
