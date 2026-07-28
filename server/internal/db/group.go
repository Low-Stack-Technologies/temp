package db

import (
	"context"
	"database/sql"
	"time"
)

type Group struct {
	ID          string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Finalized   bool
	ArchiveName *string
}

type GroupFile struct {
	ID       string
	GroupID  string
	Filename string
	Size     int64
}

func GetGroup(ctx context.Context, id string) (Group, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	row := databaseConnection.QueryRowContext(ctx, `SELECT id, created_at, expires_at, finalized, archive_name FROM upload_groups WHERE id = ? AND expires_at > CURRENT_TIMESTAMP`, id)
	return scanGroup(row)
}

func GetExpiredGroups(ctx context.Context) ([]Group, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	rows, err := databaseConnection.QueryContext(ctx, `SELECT id, created_at, expires_at, finalized, archive_name FROM upload_groups WHERE expires_at < CURRENT_TIMESTAMP ORDER BY expires_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []Group
	for rows.Next() {
		var group Group
		var finalized int
		if err := rows.Scan(&group.ID, &group.CreatedAt, &group.ExpiresAt, &finalized, &group.ArchiveName); err != nil {
			return nil, err
		}
		group.Finalized = finalized != 0
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func FinalizeGroup(ctx context.Context, id string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	_, err := databaseConnection.ExecContext(ctx, `UPDATE upload_groups SET finalized = 1 WHERE id = ?`, id)
	return err
}

func CreateGroup(ctx context.Context, id string, expiration time.Duration, archiveName *string) (Group, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	row := databaseConnection.QueryRowContext(ctx, `INSERT INTO upload_groups (id, expires_at, archive_name) VALUES (?, datetime('now', '+' || ? || ' minutes'), ?) RETURNING id, created_at, expires_at, finalized, archive_name`, id, int64(expiration.Minutes()), archiveName)
	return scanGroup(row)
}

func CreateGroupFile(ctx context.Context, file GroupFile) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	_, err := databaseConnection.ExecContext(ctx, `INSERT INTO files (id, filename, size, group_id, expires_at) SELECT ?, ?, ?, ?, expires_at FROM upload_groups WHERE id = ?`, file.ID, file.Filename, file.Size, file.GroupID, file.GroupID)
	return err
}

func GetGroupFiles(ctx context.Context, groupID string) ([]GroupFile, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	rows, err := databaseConnection.QueryContext(ctx, `SELECT id, group_id, filename, size FROM files WHERE group_id = ? ORDER BY rowid`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []GroupFile
	for rows.Next() {
		var file GroupFile
		if err := rows.Scan(&file.ID, &file.GroupID, &file.Filename, &file.Size); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

func GetGroupFile(ctx context.Context, groupID, fileID string) (GroupFile, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	row := databaseConnection.QueryRowContext(ctx, `SELECT id, group_id, filename, size FROM files WHERE group_id = ? AND id = ?`, groupID, fileID)
	var file GroupFile
	err := row.Scan(&file.ID, &file.GroupID, &file.Filename, &file.Size)
	return file, err
}

func DeleteGroup(ctx context.Context, id string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if _, err := databaseConnection.ExecContext(ctx, `DELETE FROM files WHERE group_id = ?`, id); err != nil {
		return err
	}
	_, err := databaseConnection.ExecContext(ctx, `DELETE FROM upload_groups WHERE id = ?`, id)
	return err
}

func scanGroup(row *sql.Row) (Group, error) {
	var group Group
	var finalized int
	err := row.Scan(&group.ID, &group.CreatedAt, &group.ExpiresAt, &finalized, &group.ArchiveName)
	group.Finalized = finalized != 0
	return group, err
}
