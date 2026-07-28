-- SQLite

-- +migrate Up
CREATE TABLE upload_groups (
  id TEXT PRIMARY KEY,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME NOT NULL,
  finalized INTEGER NOT NULL DEFAULT 0
);

ALTER TABLE files ADD COLUMN group_id TEXT REFERENCES upload_groups(id) ON DELETE CASCADE;
ALTER TABLE files ADD COLUMN size INTEGER NOT NULL DEFAULT 0;
CREATE INDEX files_group_id_idx ON files(group_id);

-- +migrate Down
DROP INDEX files_group_id_idx;
DROP TABLE upload_groups;
