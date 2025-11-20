package db

import (
	"database/sql"
	"fmt"
	"time"
)

// FileRecord represents a file metadata record in the database
type FileRecord struct {
	ID           string
	Filename     string
	Size         int64
	MimeType     string
	StoragePath  string
	UploadMethod string
	UploadedAt   time.Time
	TTLSeconds   int
	ExpiresAt    time.Time
	Checksum     string
}

// Repository provides database operations for file records
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateFileRecord inserts a new file record with TTL
func (r *Repository) CreateFileRecord(record *FileRecord) error {
	query := `
		INSERT INTO files (
			id, filename, size, mime_type, storage_path, 
			upload_method, uploaded_at, ttl_seconds, expires_at, checksum
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(
		query,
		record.ID,
		record.Filename,
		record.Size,
		record.MimeType,
		record.StoragePath,
		record.UploadMethod,
		record.UploadedAt,
		record.TTLSeconds,
		record.ExpiresAt,
		record.Checksum,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create file record: %w", err)
	}
	
	return nil
}

// GetFileByID retrieves a file record by ID
func (r *Repository) GetFileByID(id string) (*FileRecord, error) {
	query := `
		SELECT id, filename, size, mime_type, storage_path, 
		       upload_method, uploaded_at, ttl_seconds, expires_at, checksum
		FROM files
		WHERE id = ?
	`
	
	record := &FileRecord{}
	err := r.db.QueryRow(query, id).Scan(
		&record.ID,
		&record.Filename,
		&record.Size,
		&record.MimeType,
		&record.StoragePath,
		&record.UploadMethod,
		&record.UploadedAt,
		&record.TTLSeconds,
		&record.ExpiresAt,
		&record.Checksum,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file record: %w", err)
	}
	
	return record, nil
}

// ListFiles retrieves all file records
func (r *Repository) ListFiles(limit, offset int) ([]*FileRecord, error) {
	query := `
		SELECT id, filename, size, mime_type, storage_path, 
		       upload_method, uploaded_at, ttl_seconds, expires_at, checksum
		FROM files
		ORDER BY uploaded_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer rows.Close()
	
	var records []*FileRecord
	for rows.Next() {
		record := &FileRecord{}
		err := rows.Scan(
			&record.ID,
			&record.Filename,
			&record.Size,
			&record.MimeType,
			&record.StoragePath,
			&record.UploadMethod,
			&record.UploadedAt,
			&record.TTLSeconds,
			&record.ExpiresAt,
			&record.Checksum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file record: %w", err)
		}
		records = append(records, record)
	}
	
	return records, nil
}

// FindExpiredFiles queries files where expires_at <= current time
func (r *Repository) FindExpiredFiles() ([]*FileRecord, error) {
	query := `
		SELECT id, filename, size, mime_type, storage_path, 
		       upload_method, uploaded_at, ttl_seconds, expires_at, checksum
		FROM files
		WHERE expires_at <= ?
	`
	
	rows, err := r.db.Query(query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to find expired files: %w", err)
	}
	defer rows.Close()
	
	var records []*FileRecord
	for rows.Next() {
		record := &FileRecord{}
		err := rows.Scan(
			&record.ID,
			&record.Filename,
			&record.Size,
			&record.MimeType,
			&record.StoragePath,
			&record.UploadMethod,
			&record.UploadedAt,
			&record.TTLSeconds,
			&record.ExpiresAt,
			&record.Checksum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expired file record: %w", err)
		}
		records = append(records, record)
	}
	
	return records, nil
}

// DeleteFileRecord deletes a file record by ID
func (r *Repository) DeleteFileRecord(id string) error {
	query := `DELETE FROM files WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("file record not found")
	}
	
	return nil
}
