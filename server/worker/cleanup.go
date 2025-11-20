package worker

import (
	"context"
	"log"
	"time"

	"github.com/Low-Stack-Technologies/temp/server/config"
	"github.com/Low-Stack-Technologies/temp/server/db"
	"github.com/Low-Stack-Technologies/temp/server/storage"
)

// CleanupWorker periodically removes expired files
type CleanupWorker struct {
	cfg     *config.Config
	repo    *db.Repository
	storage *storage.Storage
	stopCh  chan struct{}
}

// NewCleanupWorker creates a new cleanup worker
func NewCleanupWorker(cfg *config.Config, repo *db.Repository, stor *storage.Storage) *CleanupWorker {
	return &CleanupWorker{
		cfg:     cfg,
		repo:    repo,
		storage: stor,
		stopCh:  make(chan struct{}),
	}
}

// Start begins the cleanup worker
func (w *CleanupWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(w.cfg.CleanupIntervalSeconds) * time.Second)
	defer ticker.Stop()
	
	log.Printf("Cleanup worker started (interval: %ds)", w.cfg.CleanupIntervalSeconds)
	
	// Run cleanup immediately on start
	w.cleanup()
	
	for {
		select {
		case <-ticker.C:
			w.cleanup()
		case <-w.stopCh:
			log.Println("Cleanup worker stopped")
			return
		case <-ctx.Done():
			log.Println("Cleanup worker stopped (context cancelled)")
			return
		}
	}
}

// Stop stops the cleanup worker
func (w *CleanupWorker) Stop() {
	close(w.stopCh)
}

// cleanup removes expired files
func (w *CleanupWorker) cleanup() {
	// Find expired files
	expiredFiles, err := w.repo.FindExpiredFiles()
	if err != nil {
		log.Printf("Error finding expired files: %v", err)
		return
	}
	
	if len(expiredFiles) == 0 {
		return
	}
	
	log.Printf("Found %d expired files to clean up", len(expiredFiles))
	
	// Delete each expired file
	deletedCount := 0
	for _, file := range expiredFiles {
		// Delete from filesystem
		if err := w.storage.DeleteFile(file.StoragePath); err != nil {
			log.Printf("Error deleting file %s from storage: %v", file.ID, err)
			// Continue with database deletion even if file deletion fails
		}
		
		// Delete from database
		if err := w.repo.DeleteFileRecord(file.ID); err != nil {
			log.Printf("Error deleting file record %s: %v", file.ID, err)
			continue
		}
		
		deletedCount++
	}
	
	log.Printf("Cleaned up %d/%d expired files", deletedCount, len(expiredFiles))
}
